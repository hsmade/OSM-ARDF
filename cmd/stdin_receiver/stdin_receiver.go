// Runs the UDP receiver that consumes ARDF measurements from the network clients
package main

import (
	"flag"
	"github.com/hsmade/OSM-ARDF/pkg/database"
	"github.com/hsmade/OSM-ARDF/pkg/receivers/stdin"
	"log"
	"os"
)

var (
	dbHost     = flag.String("database-host", "localhost", "TimescaleDB hostname")
	dbPort     = flag.Uint("database-port", 5432, "TimescaleDB port")
	dbUsername = flag.String("database-username", "postgres", "TimescaleDB username")
	dbPassword = flag.String("database-password", "postgres", "TimescaleDB password")
)

func main() {
	receiver := stdin.Receiver{Database: &database.TimescaleDB{
		Host:     *dbHost,
		Port:     uint16(*dbPort),
		Username: *dbUsername,
		Password: *dbPassword,
	}}

	log.Fatal(receiver.Start(os.Stdin))
}
