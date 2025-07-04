// internal/handler/upload_handler.go
package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"djj-inventory-system/internal/model/dto"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	// 上传根目录，本地存储时用
	UploadDir string
	// 对外访问的基础 URL，比如 "https://yourdomain.com/uploads"
	BaseURL string
}

func NewUploadHandler(rg *gin.RouterGroup, uploadDir, baseURL string) {
	h := &UploadHandler{
		UploadDir: uploadDir,
		BaseURL:   baseURL,
	}
	rg.POST("/upload", h.UploadFile)
	rg.POST("/upload/multiple", h.UploadFiles)
	rg.DELETE("/upload/delete", h.DeleteFile)
}

// UploadFile  单文件上传： field="file", form field "folder" 可选
func (h *UploadHandler) UploadFile(c *gin.Context) {
	// 1. 取文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	folder := c.DefaultPostForm("folder", "products")

	// 2. 确保目录存在
	dir := filepath.Join(h.UploadDir, folder)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create folder"})
		return
	}

	// 3. 保存文件，使用 timestamp 前缀防重名
	dstName := fmt.Sprintf("%d_%s", unixMillis(), sanitize(file.Filename))
	dstPath := filepath.Join(dir, dstName)
	if err := c.SaveUploadedFile(file, dstPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot save file"})
		return
	}

	// 4. 返回前端可访问的 URL
	url := strings.TrimRight(h.BaseURL, "/") + "/" + folder + "/" + dstName
	resp := dto.UploadResponse{
		Success:  true,
		URL:      url,
		Filename: file.Filename,
		Size:     file.Size,
	}
	c.JSON(http.StatusOK, resp)
}

// UploadFiles 多文件上传： field="files"
func (h *UploadHandler) UploadFiles(c *gin.Context) {
	folder := c.DefaultPostForm("folder", "products")
	dir := filepath.Join(h.UploadDir, folder)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create folder"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}

	files := form.File["files"]
	var results []dto.UploadResponse
	for _, file := range files {
		dstName := fmt.Sprintf("%d_%s", unixMillis(), sanitize(file.Filename))
		dstPath := filepath.Join(dir, dstName)
		if err := c.SaveUploadedFile(file, dstPath); err != nil {
			// 单个文件失败也记录
			results = append(results, dto.UploadResponse{
				Success: false,
				Message: fmt.Sprintf("failed to save %s", file.Filename),
			})
			continue
		}
		url := strings.TrimRight("uploads", "/") + "/" + folder + "/" + dstName
		results = append(results, dto.UploadResponse{
			Success:  true,
			URL:      url,
			Filename: file.Filename,
			Size:     file.Size,
		})
	}

	c.JSON(http.StatusOK, results)
}

// DeleteFile 删除接口，前端传 { fileUrl: string }
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	var body struct {
		FileUrl string `json:"fileUrl"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.FileUrl == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileUrl is required"})
		return
	}

	// 从 URL 中反推文件在磁盘的路径
	// 假设 BaseURL + "/<folder>/<filename>"
	rel := strings.TrimPrefix(body.FileUrl, strings.TrimRight(h.BaseURL, "/")+"/")
	fullPath := filepath.Join(h.UploadDir, rel)

	if err := os.Remove(fullPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot delete file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "file deleted"})
}

// sanitize 简单清理文件名（去空格等）
func sanitize(name string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '/' || r == '\\' {
			return '_'
		}
		return r
	}, name)
}

// unixMillis 返回当前毫秒级时间戳
func unixMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
