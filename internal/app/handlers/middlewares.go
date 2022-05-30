package handlers

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/romm80/shortener.git/internal/app/server"
	"github.com/romm80/shortener.git/internal/app/service"
)

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (r *gzipWriter) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

func GzipMiddleware(c *gin.Context) {
	if strings.Contains(c.Request.Header.Get("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		defer gz.Close()
		c.Request.Body = gz
	}
	if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
		c.Header("Content-Encoding", "gzip")
		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()
		c.Writer = &gzipWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}
	}
	c.Next()
}

func (s *Shortener) AuthMiddleware(c *gin.Context) {
	var userID uint64

	cookie, err := c.Cookie("userid")
	if err != nil || !service.ValidUserID(cookie, &userID) {
		if userID, err = s.Storage.NewUser(); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		signedID, err := service.SignUserID(userID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.SetCookie("userid", signedID, 0, "/", server.Cfg.Domain, false, true)
	}

	c.Set("userid", userID)
	c.Next()
}
