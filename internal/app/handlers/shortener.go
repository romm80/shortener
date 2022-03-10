package handlers

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/romm80/shortener.git/internal/app/repositories"
	"github.com/romm80/shortener.git/internal/app/repositories/mapstorage"
	"github.com/romm80/shortener.git/internal/app/server"
	"io/ioutil"
	"net/http"
	"strings"
)

type Shortener struct {
	Router  *gin.Engine
	Storage repositories.Shortener
}

type OriginURL struct {
	URL string `json:"url"`
}

type ShortURL struct {
	Result string `json:"result"`
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (r *gzipWriter) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

func New() (*Shortener, error) {
	r := &Shortener{}
	s, err := mapstorage.New()
	if err != nil {
		return nil, err
	}
	r.Storage = s

	r.Router = gin.Default()
	r.Router.Use(r.AuthMiddleware, r.GzipMiddleware)
	r.Router.GET("/:id", r.Get)
	r.Router.POST("/", r.Add)
	r.Router.POST("/api/shorten", r.AddJSON)
	r.Router.GET("/api/user/urls", r.GetUserURLs)

	return r, nil
}

func (s *Shortener) Add(c *gin.Context) {
	link, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	userID := c.GetUint64("userid")
	id, err := s.Storage.Add(string(link), userID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusCreated, "%s/%s", server.Cfg.BaseURL, id)
}

func (s *Shortener) AddJSON(c *gin.Context) {
	var url OriginURL
	if err := json.NewDecoder(c.Request.Body).Decode(&url); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if url.URL == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id, err := s.Storage.Add(url.URL, c.GetUint64("userid"))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	res := ShortURL{Result: fmt.Sprintf("%s/%s", server.Cfg.BaseURL, id)}
	c.JSON(http.StatusCreated, res)
}

func (s *Shortener) Get(c *gin.Context) {
	id := c.Param("id")
	link, err := s.Storage.Get(id)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, link)
}

func (s *Shortener) GetUserURLs(c *gin.Context) {
	userID := c.GetUint64("userid")
	res := s.Storage.GetUserURLs(userID)
	if len(res) == 0 {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Shortener) GzipMiddleware(c *gin.Context) {
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
	if err != nil || invalidUserID(cookie, &userID) || !s.Storage.CheckUserID(userID) {
		userID = s.Storage.NewUser()
		signedId, err := signUserID(userID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.SetCookie("userid", signedId, 60*60*24, "/", server.Cfg.Domain, false, true)
	}

	c.Set("userid", userID)
	c.Next()
}

func signUserID(id uint64) (string, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, id)

	h := hmac.New(sha256.New, server.Cfg.SecretKey)
	if _, err := h.Write(buf); err != nil {
		return "", err
	}
	res := h.Sum(nil)
	return hex.EncodeToString(append(buf, res...)), nil
}

func invalidUserID(src string, userID *uint64) bool {
	data, err := hex.DecodeString(src)
	if err != nil {
		return true
	}

	*userID = binary.BigEndian.Uint64(data[:8])
	h := hmac.New(sha256.New, server.Cfg.SecretKey)
	h.Write(data[:8])
	sign := h.Sum(nil)

	return !hmac.Equal(sign, data[8:])
}
