package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/romm80/shortener.git/internal/app/repositories"
	"github.com/romm80/shortener.git/internal/app/repositories/mapstorage"
	"github.com/romm80/shortener.git/internal/app/server"
	"io/ioutil"
	"net/http"
)

type Shortener struct {
	Router  *gin.Engine
	Storage repositories.Shortener
}

func New() *Shortener {
	r := &Shortener{}
	r.Storage = mapstorage.New()

	gin.DefaultWriter = ioutil.Discard
	gin.SetMode(gin.ReleaseMode)
	r.Router = gin.Default()
	r.Router.POST("/", r.Add)
	r.Router.GET("/:id", r.Get)

	return r
}

func (s *Shortener) Add(c *gin.Context) {
	link, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	id := s.Storage.Add(string(link))
	c.String(http.StatusCreated, "http://%v:%v/%v", server.Cfg.Addr, server.Cfg.Port, id)

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
