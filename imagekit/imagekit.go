package imagekit

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path"

	"github.com/PabloPavan/jaiu/imagekit/config"
	kitimage "github.com/PabloPavan/jaiu/imagekit/image"
	"github.com/PabloPavan/jaiu/imagekit/outbox"
	"github.com/PabloPavan/jaiu/imagekit/processor"
	"github.com/PabloPavan/jaiu/imagekit/queue"
	"github.com/PabloPavan/jaiu/imagekit/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

const (
	defaultQueueName   = "imagekit:queue"
	defaultOriginalKey = "original.jpg"
)

type Enqueuer interface {
	Enqueue(ctx context.Context, msg queue.Message) error
}

type TxBeginner interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
}

type Kit struct {
	storage      storage.ObjectStorage
	queue        queue.Queue
	enqueuer     Enqueuer
	processor    *processor.Processor
	originalName string
	txBeginner   TxBeginner
}

func New(ctx context.Context, cfg config.Config) (*Kit, error) {
	queueName := cfg.QueueName
	if queueName == "" {
		queueName = defaultQueueName
	}
	originalName := cfg.OriginalKey
	if originalName == "" {
		originalName = defaultOriginalKey
	}

	var objectStorage storage.ObjectStorage
	switch cfg.StorageType {
	case "", config.StorageR2:
		r2Storage, err := storage.NewR2Storage(ctx, cfg.R2)
		if err != nil {
			return nil, err
		}
		objectStorage = r2Storage
	case config.StorageLocal:
		localStorage, err := storage.NewLocalStorage(cfg.LocalDir)
		if err != nil {
			return nil, err
		}
		objectStorage = localStorage
	default:
		return nil, errors.New("unsupported storage type")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	redisQueue := queue.NewRedisQueue(redisClient, queueName)

	sizes := make([]kitimage.Size, 0, len(cfg.Sizes))
	for _, size := range cfg.Sizes {
		sizes = append(sizes, kitimage.Size{
			Name:   size.Name,
			Width:  size.Width,
			Height: size.Height,
		})
	}

	proc := &processor.Processor{
		Storage:      objectStorage,
		Sizes:        sizes,
		OriginalName: originalName,
	}

	return &Kit{
		storage:      objectStorage,
		queue:        redisQueue,
		enqueuer:     redisQueue,
		processor:    proc,
		originalName: originalName,
	}, nil
}

func NewWithDeps(storage storage.ObjectStorage, queue queue.Queue, sizes []kitimage.Size, originalName string) *Kit {
	return NewWithEnqueuer(storage, queue, queue, sizes, originalName)
}

func NewWithEnqueuer(storage storage.ObjectStorage, queue queue.Queue, enqueuer Enqueuer, sizes []kitimage.Size, originalName string) *Kit {
	if originalName == "" {
		originalName = defaultOriginalKey
	}
	proc := &processor.Processor{
		Storage:      storage,
		Sizes:        sizes,
		OriginalName: originalName,
	}
	if enqueuer == nil {
		enqueuer = queue
	}
	return &Kit{
		storage:      storage,
		queue:        queue,
		enqueuer:     enqueuer,
		processor:    proc,
		originalName: originalName,
	}
}

func (k *Kit) SetEnqueuer(enqueuer Enqueuer) {
	if enqueuer != nil {
		k.enqueuer = enqueuer
	}
}

func (k *Kit) SetTxBeginner(txBeginner TxBeginner) {
	if txBeginner != nil {
		k.txBeginner = txBeginner
	}
}

func (k *Kit) UploadImage(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	if k.txBeginner == nil {
		return k.uploadImage(ctx, file, header)
	}
	tx, err := k.txBeginner.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", err
	}
	ctx = outbox.ContextWithTx(ctx, tx)

	objectKey, err := k.uploadImage(ctx, file, header)
	if err != nil {
		_ = tx.Rollback(ctx)
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return objectKey, nil
}

func (k *Kit) uploadImage(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", errors.New("file is required")
	}
	if k.enqueuer == nil {
		return "", errors.New("enqueuer is required")
	}

	objectKey := uuid.NewString()
	contentType := ""
	if header != nil {
		contentType = header.Header.Get("Content-Type")
	}

	reader := io.Reader(file)
	if contentType == "" {
		var err error
		contentType, reader, err = sniffContentType(file)
		if err != nil {
			return "", err
		}
	}

	key := path.Join(objectKey, k.originalName)
	if err := k.storage.Put(ctx, key, reader, contentType); err != nil {
		return "", err
	}

	if err := k.enqueuer.Enqueue(ctx, queue.Message{ObjectKey: objectKey}); err != nil {
		return "", err
	}

	return objectKey, nil
}

func (k *Kit) StartWorkers(ctx context.Context, n int) error {
	if n <= 0 {
		return errors.New("worker count must be greater than zero")
	}
	if k.queue == nil {
		return errors.New("queue is required for workers")
	}

	group, ctx := errgroup.WithContext(ctx)
	for i := 0; i < n; i++ {
		worker := &processor.Worker{
			Queue:     k.queue,
			Processor: k.processor,
		}
		group.Go(func() error {
			return worker.Run(ctx)
		})
	}

	return group.Wait()
}

func sniffContentType(r io.Reader) (string, io.Reader, error) {
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", nil, err
	}
	if n == 0 {
		return "", nil, errors.New("empty file")
	}
	contentType := http.DetectContentType(buf[:n])
	return contentType, io.MultiReader(bytes.NewReader(buf[:n]), r), nil
}
