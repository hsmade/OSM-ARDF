package web

import (
	"github.com/matryer/is"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRootRedirect(t *testing.T) {
	Is := is.New(t)
	srv := NewServer("postgresql://postgres:postgres@localhost:5432/postgres")
	req, err := http.NewRequest("GET", "/", nil)
	Is.NoErr(err)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)
	Is.Equal(w.Result().StatusCode, http.StatusMovedPermanently)
	Is.Equal(w.Header().Get("location"), "/app/")
}
