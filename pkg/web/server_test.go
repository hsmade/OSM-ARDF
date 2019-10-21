package web

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/ory/dockertest"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

var (
	dbResource *dockertest.Resource
	dockerPort int
	db         *pgx.Conn
)

// Setup a docker container with postgres before running the tests for databases
func TestMain(m *testing.M) {
	var err error
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	initSql, err := filepath.Abs(dir + "../../../scripts/init-gis.sql")
	if err != nil {
		log.Fatal(err)
	}

	options := &dockertest.RunOptions{
		Repository: "timescale/timescaledb-postgis",
		Tag:        "1.4.2-pg11",
		Env:        []string{"POSTGRES_PASSWORD=postgres"},
		Mounts:     []string{initSql + ":/docker-entrypoint-initdb.d/init-gis.sql"},
	}
	dbResource, err = pool.RunWithOptions(options)
	if err != nil {
		log.Fatalf("Could not start dbResource: %s", err)
	}
	err = dbResource.Expire(60)
	if err != nil {
		log.Fatalf("Could not set expiration for docker container: %s", err)
	}

	if err = pool.Retry(func() error {
		var err error
		db, err = pgx.Connect(
			context.Background(),
			fmt.Sprintf("postgresql://postgres:postgres@localhost:%s/postgres?sslmode=disable", dbResource.GetPort("5432/tcp")))
		if err != nil {
			return err
		}

		_, err = db.Exec(context.Background(), "SELECT * FROM doppler;")
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	dockerPort, err = strconv.Atoi(dbResource.GetPort("5432/tcp"))
	if err != nil {
		log.Fatalf("Failed to get port for docker container: %e", err)
	}

	exit := m.Run()

	err = pool.Purge(dbResource)
	os.Exit(exit)
}

func TestNewServer(t *testing.T) {
	s := NewServer(fmt.Sprintf("postgresql://postgres:postgres@localhost:%d/postgres", dockerPort))

	if _, err := s.db.GetPositions(time.Second); err != nil {
		t.Errorf("server did not connect to DB? Got error during query: %e", err)
	}

	r := s.router.Routes()
	if len(r) < 1 {
		t.Errorf("no routes loaded")
	}
}
