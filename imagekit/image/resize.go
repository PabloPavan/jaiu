package image

import (
	"bytes"
	"errors"
	stdimage "image"
	"image/jpeg"

	"github.com/disintegration/imaging"
)

type Size struct {
	Name   string
	Width  int
	Height int
}

func ResizeVariants(src stdimage.Image, sizes []Size) (map[string][]byte, error) {
	variants := make(map[string][]byte, len(sizes))

	for _, size := range sizes {
		if size.Name == "" {
			return nil, errors.New("image size name is required")
		}
		if size.Width <= 0 && size.Height <= 0 {
			return nil, errors.New("image size must define width or height")
		}

		var dst *stdimage.NRGBA
		if size.Width > 0 && size.Height > 0 {
			dst = imaging.Fit(src, size.Width, size.Height, imaging.Lanczos)
		} else {
			dst = imaging.Resize(src, size.Width, size.Height, imaging.Lanczos)
		}

		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 85}); err != nil {
			return nil, err
		}
		variants[size.Name] = buf.Bytes()
	}

	return variants, nil
}
