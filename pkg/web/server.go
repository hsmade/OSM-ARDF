package web

import (
	"github.com/apex/log"
	"github.com/gin-gonic/gin"
	"github.com/hsmade/OSM-ARDF/pkg/database"
	"net/http"
)

type server struct {
	router *gin.Engine
	db     database.Database
}

func NewServer(databaseURL string) *server {
	s := server{router: gin.Default()}
	s.routes()
	s.db = database.New(databaseURL)
	err := s.db.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %e", err)
	} else {
		log.Info("connected to database")
	}
	return &s
}

func (s *server) Serve(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
