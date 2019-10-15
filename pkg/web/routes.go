package web

import (
	"github.com/gin-gonic/gin"
	"github.com/hsmade/OSM-ARDF/web"
)

func (s *server) routes() {
	s.router.StaticFS("/app", web.Assets)
	s.router.GET("/", s.redirectRoot)
}

func (s *server) redirectRoot(c *gin.Context) {
	c.Request.URL.Path = "/app"
	s.router.HandleContext(c)
}
