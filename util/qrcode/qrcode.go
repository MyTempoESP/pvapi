package qrcode

import (
	"encoding/base64"
	"fmt"

	goqr "github.com/skip2/go-qrcode"
)

const (
	mimeType = "image/png"
	size     = 256
)

// Encode returns a base64-encoded data URI string representing
// the QR code PNG for the given text.
func Encode(text string) (string, error) {
	png, err := goqr.Encode(text, goqr.Medium, size)
	if err != nil {
		return "", fmt.Errorf("encode qr: %w", err)
	}

	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(png))
	return dataURI, nil
}
