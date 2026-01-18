package handlers

import (
	"context"
	"mime/multipart"
	"net/http"
)

type ImageService interface {
	UploadImage(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
	Handler() http.Handler
}

type ImageConfig struct {
	ImageService ImageService
	BaseURL      string
	OriginalKey  string
	FieldName    string
	MaxMemory    int64
}

func (h *Handler) SetImageConfig(config ImageConfig) {
	if config.BaseURL == "" {
		config.BaseURL = "/images"
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

func (h *Handler) ImageHandler() http.Handler {
	{
		if h.images.ImageService == nil {
			return nil
		}
		return h.images.ImageService.Handler()
	}
}
