// Runs the webserver that is used by the web-app to fetch points, lines and heat maps

package main

import (
	"github.com/hsmade/OSM-ARDF/pkg/web"
	"log"
	"os"
)

func main() {
	s := web.NewServer(os.Getenv("DATABASE"))
	log.Fatal(s.Serve(":8083"))
}
