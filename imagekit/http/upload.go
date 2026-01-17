package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/PabloPavan/jaiu/imagekit"
)

type UploadHandler struct {
	Kit       *imagekit.Kit
	FieldName string
	MaxMemory int64
}

func (h *UploadHandler) ServeHTTP(w nethttp.ResponseWriter, r *nethttp.Request) {
	if h.Kit == nil {
		w.WriteHeader(nethttp.StatusInternalServerError)
		return
	}
	if r.Method != nethttp.MethodPost {
		w.WriteHeader(nethttp.StatusMethodNotAllowed)
		return
	}

	field := h.FieldName
	if field == "" {
		field = "file"
	}
	maxMemory := h.MaxMemory
	if maxMemory == 0 {
		maxMemory = 32 << 20
	}

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		w.WriteHeader(nethttp.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile(field)
	if err != nil {
		w.WriteHeader(nethttp.StatusBadRequest)
		return
	}
	defer file.Close()

	objectKey, err := h.Kit.UploadImage(r.Context(), file, header)
	if err != nil {
		w.WriteHeader(nethttp.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"object_key": objectKey,
	})
}
