package web

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/hsmade/OSM-ARDF/pkg/datastructures"
	"github.com/paulmach/go.geojson"
	"log"
	"strconv"
	"time"
)

func (s *server) handlePostions() gin.HandlerFunc {
	// FIXME since parameter

	return func(c *gin.Context) {
		seconds := c.Query("seconds")
		since, err := strconv.Atoi(seconds)
		if err != nil {
			_ = c.AbortWithError(500, errors.New("seconds must be a number"))
			return
		}
		positions, _ := s.db.GetPositions(time.Duration(since) * time.Second) // FIXME handle error
		log.Printf("got %d positions", len(positions))
		c.String(200, string(formatPositions(positions)))
	}
}

func formatPositions(positions []*datastructures.Position) []byte { // FIXME handle error
	fc := geojson.NewFeatureCollection()
	for _, point := range positions {
		pointFeature := geojson.NewPointFeature([]float64{point.Longitude, point.Latitude})
		pointFeature.Properties = map[string] interface{}{"id": point.Station+point.Timestamp.String()}
		fc.AddFeature(pointFeature)
	}
	rawJSON, err := fc.MarshalJSON()
	if err != nil {
		log.Printf("error marshalling json: %e", err)
		return []byte("error")
	}
	log.Printf("raw json: %s", string(rawJSON))
	return rawJSON
}
