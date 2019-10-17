package web

import (
	"github.com/apex/log"
	"github.com/gin-gonic/gin"
	"github.com/hsmade/OSM-ARDF/web"
)

func (s *server) routes() {
	log.Info("loading routes")
	s.router.StaticFS("/app", web.Assets)
	s.router.GET("/", s.redirectRoot)
	api := s.router.Group("api")
	api.GET("/positions", s.handlePostions())

}

func (s *server) redirectRoot(c *gin.Context) {
	c.Request.URL.Path = "/app"
	s.router.HandleContext(c)
}
