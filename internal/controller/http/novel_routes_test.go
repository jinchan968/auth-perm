package http

import (
	"testing"

	"github.com/gin-gonic/gin"

	novelHandler "auth-perm/internal/domain/novel/handler"
)

func TestRegisterNovelRoutesDoesNotPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")

	RegisterNovelRoutes(v1, func(c *gin.Context) { c.Next() }, novelHandler.NewNovelHandler(nil), nil)
}
