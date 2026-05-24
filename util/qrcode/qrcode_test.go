package qrcode

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	text := "checkout-uuid-12345"
	dataURI, err := Encode(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(dataURI, "data:image/png;base64,") {
		t.Fatalf("expected data URI prefix, got: %s", dataURI[:50])
	}

	b64 := strings.TrimPrefix(dataURI, "data:image/png;base64,")
	decoded, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatalf("failed to decode base64: %v", err)
	}

	// PNG signature
	pngSig := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	if len(decoded) < len(pngSig) {
		t.Fatal("decoded data too short for PNG")
	}
	for i, b := range pngSig {
		if decoded[i] != b {
			t.Fatalf("not a valid PNG at byte %d: expected 0x%02x, got 0x%02x", i, b, decoded[i])
		}
	}
}

func TestEncodeEmpty(t *testing.T) {
	_, err := Encode("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}
