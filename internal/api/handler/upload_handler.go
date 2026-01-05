package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// Upload BRD
func (h *UploadHandler) UploadBRD(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File tidak ditemukan. Pastikan key form-data adalah 'file'",
		})
		return
	}

	uploadDir := "./uploads"

	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.Mkdir(uploadDir, 0755)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat folder upload"})
			return
		}
	}

	filename := fmt.Sprintf("%d-%s", time.Now().Unix(), filepath.Base(file.Filename))
	dst := filepath.Join(uploadDir, filename)

	// 4. Simpan file ke folder tujuan
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file ke server"})
		return
	}

	// 5. Berikan respon sukses
	c.JSON(http.StatusOK, gin.H{
		"message":       "File berhasil diupload",
		"filename":      filename,
		"filepath":      dst,
		"original_name": file.Filename,
		"size":          file.Size,
	})
}
