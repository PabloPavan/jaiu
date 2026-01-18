package handlers

import (
	"context"
	"mime/multipart"
)

type ImageUploader interface {
	UploadImage(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
}

type ImageConfig struct {
	Uploader    ImageUploader
	BaseURL     string
	OriginalKey string
	FieldName   string
	MaxMemory   int64
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
