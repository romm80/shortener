package handlers

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/romm80/shortener.git/internal/app"
	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/repositories"
	"github.com/romm80/shortener.git/internal/app/repositories/dbpostgres"
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

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (r *gzipWriter) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

func New() (*Shortener, error) {
	r := &Shortener{}
	var err error

	switch server.Cfg.DBType {
	case server.DBMap:
		r.Storage, err = mapstorage.New()
	case server.DBPostgres:
		r.Storage, err = dbpostgres.New()
	default:
		return nil, errors.New("wrong DB type")
	}
	if err != nil {
		return nil, err
	}

	r.Router = gin.Default()
	r.Router.GET("/ping", r.PingDB)
	r.Router.Use(r.GzipMiddleware)
	r.Router.GET("/:id", r.Get)
	r.Router.Use(r.AuthMiddleware)
	r.Router.POST("/", r.Add)
	r.Router.POST("/api/shorten", r.AddJSON)
	r.Router.POST("/api/shorten/batch", r.BatchURLs)
	r.Router.GET("/api/user/urls", r.GetUserURLs)

	return r, nil
}

func (s *Shortener) Add(c *gin.Context) {
	originURL, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	urlID, err := s.Storage.Add(string(originURL), c.GetUint64("userid"))
	statusCode := http.StatusCreated
	if err != nil {
		if errors.Is(err, app.ErrConflictURLID) {
			statusCode = http.StatusConflict
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}
	c.String(statusCode, "%s/%s", server.Cfg.BaseURL, urlID)
}

func (s *Shortener) AddJSON(c *gin.Context) {
	var request models.RequestURL
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if request.URL == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	urlID, err := s.Storage.Add(request.URL, c.GetUint64("userid"))
	statusCode := http.StatusCreated
	if err != nil {
		if errors.Is(err, app.ErrConflictURLID) {
			statusCode = http.StatusConflict
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}
	res := models.ResponseURL{Result: fmt.Sprintf("%s/%s", server.Cfg.BaseURL, urlID)}
	c.JSON(statusCode, res)
}

func (s *Shortener) Get(c *gin.Context) {
	urlID := c.Param("id")
	originURL, err := s.Storage.Get(urlID)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, originURL)
}

func (s *Shortener) BatchURLs(c *gin.Context) {
	reqBatch := make([]models.RequestBatch, 0)
	if err := json.NewDecoder(c.Request.Body).Decode(&reqBatch); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	respBatch, err := s.Storage.AddBatch(reqBatch, c.GetUint64("userid"))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, respBatch)
}
func (s *Shortener) GetUserURLs(c *gin.Context) {
	userID := c.GetUint64("userid")
	res, err := s.Storage.GetUserURLs(userID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
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
	if err != nil || !s.validUserID(cookie, &userID) {
		if userID, err = s.Storage.NewUser(); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		signedID, err := signUserID(userID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.SetCookie("userid", signedID, 0, "/", server.Cfg.Domain, false, true)
	}

	c.Set("userid", userID)
	c.Next()
}

func (s *Shortener) PingDB(c *gin.Context) {
	if err := s.Storage.Ping(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
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

func (s *Shortener) validUserID(src string, userID *uint64) bool {
	data, err := hex.DecodeString(src)
	if err != nil {
		return false
	}

	*userID = binary.BigEndian.Uint64(data[:8])
	h := hmac.New(sha256.New, server.Cfg.SecretKey)
	h.Write(data[:8])
	sign := h.Sum(nil)

	return hmac.Equal(sign, data[8:])
}
