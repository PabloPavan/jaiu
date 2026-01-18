package imagekithttp

import (
	"io"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/PabloPavan/jaiu/imagekit/storage"
)

type Handler struct {
	storage storage.ObjectStorage
}

func NewHandler(objectStorage storage.ObjectStorage) http.Handler {
	if objectStorage == nil {
		return nil
	}
	return &Handler{storage: objectStorage}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clean := path.Clean("/" + strings.TrimSpace(r.URL.Path))
	parts := strings.Split(strings.TrimPrefix(clean, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.NotFound(w, r)
		return
	}
	key := path.Join(parts[0], parts[1]+".jpg")

	reader, err := h.storage.Get(r.Context(), key)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer reader.Close()

	contentType := mime.TypeByExtension(path.Ext(key))
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	_, _ = io.Copy(w, reader)
}
