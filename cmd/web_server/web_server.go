// Runs the webserver that is used by the web-app to fetch points, lines and heat maps

package main

import (
	"github.com/hsmade/OSM-ARDF/pkg/web"
	"log"
)

func main() {
	s := web.NewServer()
	log.Fatal(s.Serve(":7070"))
}
