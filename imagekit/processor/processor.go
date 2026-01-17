package processor

import (
	"bytes"
	"context"
	"errors"
	"path"

	stdimage "image"

	kitimage "github.com/PabloPavan/jaiu/imagekit/image"
	"github.com/PabloPavan/jaiu/imagekit/queue"
	"github.com/PabloPavan/jaiu/imagekit/storage"
)

type Processor struct {
	Storage      storage.ObjectStorage
	Sizes        []kitimage.Size
	OriginalName string
}

func (p *Processor) Process(ctx context.Context, msg queue.Message) error {
	originalKey := path.Join(msg.ObjectKey, p.OriginalName)
	reader, err := p.Storage.Get(ctx, originalKey)
	if err != nil {
		return err
	}
	defer reader.Close()

	img, _, err := stdimage.Decode(reader)
	if err != nil {
		return err
	}

	variants, err := kitimage.ResizeVariants(img, p.Sizes)
	if err != nil {
		return err
	}

	for name, data := range variants {
		key := path.Join(msg.ObjectKey, name+".jpg")
		if err := p.Storage.Put(ctx, key, bytes.NewReader(data), "image/jpeg"); err != nil {
			return err
		}
	}

	return nil
}

type Worker struct {
	Queue     queue.Queue
	Processor *Processor
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		msg, err := w.Queue.Dequeue(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return ctx.Err()
			}
			return err
		}
		if err := w.Processor.Process(ctx, msg); err != nil {
			return err
		}
	}
}
