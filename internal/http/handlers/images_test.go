package handlers

import (
	"context"
	"mime/multipart"
	"net/http"
	"testing"
)

type fakeImageService struct {
	handler http.Handler
}

func (f *fakeImageService) UploadImage(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	return "", nil
}

func (f *fakeImageService) Handler() http.Handler {
	return f.handler
}

// Testa defaults de configuracao de imagens.
func TestSetImageConfigDefaults(t *testing.T) {
	h := &Handler{}
	h.SetImageConfig(ImageConfig{})

	if h.images.BaseURL != "/images" {
		t.Fatalf("expected base url /images, got %q", h.images.BaseURL)
	}
	if h.images.OriginalKey != "original.jpg" {
		t.Fatalf("expected original key, got %q", h.images.OriginalKey)
	}
	if h.images.FieldName != "file" {
		t.Fatalf("expected field name file, got %q", h.images.FieldName)
	}
	if h.images.MaxMemory == 0 {
		t.Fatal("expected max memory to be set")
	}
}

// Testa o retorno do handler de imagens quando o servico nao existe.
func TestImageHandlerNilService(t *testing.T) {
	h := &Handler{}
	if got := h.ImageHandler(); got != nil {
		t.Fatal("expected nil image handler when service is nil")
	}
}

// Testa o retorno do handler de imagens quando o servico existe.
func TestImageHandlerWithService(t *testing.T) {
	h := &Handler{}
	expected := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h.SetImageConfig(ImageConfig{ImageService: &fakeImageService{handler: expected}})

	if got := h.ImageHandler(); got == nil {
		t.Fatal("expected image handler when service exists")
	}
}
