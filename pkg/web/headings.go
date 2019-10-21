package web

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hsmade/OSM-ARDF/pkg/types"
	"github.com/paulmach/go.geojson"
	"log"
	"strconv"
	"time"
)

func (s *server) handleHeadings() gin.HandlerFunc {
	return func(c *gin.Context) {
		seconds := c.Query("seconds")
		since, err := strconv.Atoi(seconds)
		if err != nil {
			_ = c.AbortWithError(500, errors.New("seconds must be a number"))
			return
		}
		lines, err := s.db.GetLines(time.Duration(since) * time.Second)
		if err != nil {
			_ = c.AbortWithError(500, errors.New(fmt.Sprintf("unable to get lines: %e", err)))
			return
		}
		log.Printf("got %d lines", len(lines))
		c.String(200, string(formatHeadings(lines)))
	}
}

func formatHeadings(lines []*types.Line) []byte {
	fc := geojson.NewFeatureCollection()
	for _, line := range lines {
		pointFeature := geojson.NewLineStringFeature([][]float64{{line.Longitude, line.Latitude}, {line.LongitudeEnd, line.LatitudeEnd}})
		pointFeature.Properties = map[string]interface{}{"id": line.Station + line.Timestamp.String()}
		fc.AddFeature(pointFeature)
	}
	rawJSON, err := fc.MarshalJSON()
	if err != nil {
		log.Printf("error marshalling json: %e", err)
		return []byte("error marshalling into json")
	}
	log.Printf("raw json: %s", string(rawJSON))
	return rawJSON
}
