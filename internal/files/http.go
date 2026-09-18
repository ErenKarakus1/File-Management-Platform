package files

import (
	"errors"
	"net/http"

	"github.com/ErenKarakus1/File-Management-Platform/internal/auth"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service        *Service
	maxUploadBytes int64
}

func NewHandler(service *Service, maxUploadBytes int64) *Handler {
	return &Handler{
		service:        service,
		maxUploadBytes: maxUploadBytes,
	}
}

func (h *Handler) Upload(c *gin.Context) {
	user, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxUploadBytes)
	header, err := c.FormFile("file")
	if err != nil {
		if errors.As(err, new(*http.MaxBytesError)) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file is too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := h.service.Upload(c.Request.Context(), user.ID, header)
	if err != nil {
		if errors.Is(err, ErrEmptyFile) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is empty"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not upload file"})
		return
	}

	c.JSON(http.StatusCreated, file)
}

func (h *Handler) List(c *gin.Context) {
	user, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	files, err := h.service.List(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}
