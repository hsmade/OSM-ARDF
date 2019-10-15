package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type server struct {
	router *gin.Engine
}

func NewServer() *server {
	s := server{router: gin.Default()}
	s.routes()
	return &s
}

func (s *server) Serve(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
