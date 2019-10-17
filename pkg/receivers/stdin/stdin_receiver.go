package stdin

import (
	"bufio"
	"encoding/json"
	"github.com/apex/log"
	"github.com/hsmade/OSM-ARDF/pkg/database"
	"github.com/hsmade/OSM-ARDF/pkg/types"
	"io"
)

type Receiver struct {
	Database database.Database
}

func (r *Receiver) Start(reader io.Reader) error {
	err := r.Database.Connect()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bufio.NewReader(reader))
	for scanner.Scan() {
		r.process(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (r *Receiver) process(data string) {
	m := types.Measurement{}
	err := json.Unmarshal([]byte(data), &m)
	if err != nil {
		log.WithError(err).Error("Failed to parse into measurement")
		return
	}
	defer log.WithField("measurement", m).Trace("storing measurement").Stop(&err)
	log.WithField("measurement", m).Debug("Storing measurement")
	err = r.Database.Add(&m)
}
