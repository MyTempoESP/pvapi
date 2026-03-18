package productmodel

import (
	"fmt"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type URL string

// handleImage stores the image and returns the full path
func handleImage(c *gin.Context, m *multipart.FileHeader) (URL, error) {
	dst := fmt.Sprintf("./public/uploads/%s", m.Filename)
	err := c.SaveUploadedFile(m, dst)
	return URL(dst), err
}

func handleImages(c *gin.Context, ms []*multipart.FileHeader) ([]URL, error) {
	var (
		paths []URL
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
