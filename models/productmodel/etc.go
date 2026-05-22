package productmodel

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/gosimple/slug"
	"linhazero.com/pvapi/util"
)

func checkFileSanity(m *multipart.FileHeader) error {
	file, err := m.Open()
	if err != nil {
		return fmt.Errorf("failed to open file header: %w", err)
	}
	defer func() { _ = file.Close() }()

	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file bytes: %w", err)
	}

	contentType := http.DetectContentType(buff)
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("invalid file type: expected image, got %s", contentType)
	}

	return nil
}

// handleImage stores the image and returns the full path
func handleImage(
	c *gin.Context,
	m *multipart.FileHeader,
) (path string, err error) {
	err = checkFileSanity(m)

	if err != nil {
		return
	}

	uniqueid, err := uuid.NewV4()

	if err != nil {
		return
	}

	ext := filepath.Ext(m.Filename)
	head := strings.TrimSuffix(m.Filename, ext)

	slug := slug.Make(head)
	base := fmt.Sprintf("%s_%s%s", slug, uniqueid, ext)

	// TODO: put that on util too
	savePath := fmt.Sprintf("./public/uploads/%s", base)
	path, err = util.GetImageURL(base)

	if err != nil {
		return
	}

	err = c.SaveUploadedFile(m, savePath)

	return
}

func handleImages(c *gin.Context, ms []*multipart.FileHeader) ([]string, error) {
	var (
		paths []string
	)

	for _, m := range ms {
		p, err := handleImage(c, m)

		if err != nil {
			return nil, err
		}

		paths = append(paths, p)
	}

	return paths, nil
}
