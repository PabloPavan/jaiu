package view

import (
	"os"
	"strings"
)

func useCDNAssets() bool {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	switch env {
	case "", "local", "dev", "development", "test":
		return true
	default:
		return false
	}
}
