package handlers

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"path"
	"strings"

	"github.com/PabloPavan/jaiu/imagekit/outbox"
	"github.com/PabloPavan/jaiu/internal/observability"
	"github.com/jackc/pgx/v5"
)

type ImageUploader interface {
	UploadImage(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
}

type TxBeginner interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
}

type ImageConfig struct {
	Uploader    ImageUploader
	BaseURL     string
	OriginalKey string
	FieldName   string
	MaxMemory   int64
	TxBeginner  TxBeginner
}

func (h *Handler) SetImageConfig(config ImageConfig) {
	if config.BaseURL == "" {
		config.BaseURL = "/uploads"
	}
	if config.OriginalKey == "" {
		config.OriginalKey = "original.jpg"
	}
	if config.FieldName == "" {
		config.FieldName = "file"
	}
	if config.MaxMemory == 0 {
		config.MaxMemory = 32 << 20
	}
	h.images = config
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if h.images.Uploader == nil {
		http.Error(w, "upload nao configurado", http.StatusNotImplemented)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(h.images.MaxMemory); err != nil {
		http.Error(w, "erro ao ler arquivo", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile(h.images.FieldName)
	if err != nil {
		http.Error(w, "arquivo invalido", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ctx := r.Context()
	var tx pgx.Tx
	if h.images.TxBeginner != nil {
		tx, err = h.images.TxBeginner.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			observability.Logger(ctx).Error("failed to start outbox tx", "err", err)
			http.Error(w, "erro ao iniciar upload", http.StatusInternalServerError)
			return
		}
		ctx = outbox.ContextWithTx(ctx, tx)
	}

	objectKey, err := h.images.Uploader.UploadImage(ctx, file, header)
	if tx != nil {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else if commitErr := tx.Commit(ctx); commitErr != nil {
			observability.Logger(ctx).Error("failed to commit outbox tx", "err", commitErr)
			http.Error(w, "erro ao salvar imagem", http.StatusInternalServerError)
			return
		}
	}
	if err != nil {
		observability.Logger(r.Context()).Error("failed to upload image", "err", err)
		http.Error(w, "erro ao salvar imagem", http.StatusInternalServerError)
		return
	}

	baseURL := strings.TrimRight(h.images.BaseURL, "/")
	if baseURL == "" {
		baseURL = "/uploads"
	}
	objectPath := path.Join(objectKey, h.images.OriginalKey)
	url := baseURL + "/" + objectPath

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"object_key": objectKey,
		"url":        url,
	})
}
