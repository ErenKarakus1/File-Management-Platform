package server

import (
	"net/http"

	"github.com/ErenKarakus1/File-Management-Platform/internal/auth"

	"github.com/gin-gonic/gin"
)

func New(authService *auth.Service) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authHandler := auth.NewHandler(authService)

	api := router.Group("/api/v1")
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	protected := api.Group("")
	protected.Use(auth.Middleware(authService))
	protected.GET("/auth/me", auth.Me)

	return router
}
