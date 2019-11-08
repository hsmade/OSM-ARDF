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

func (s *server) handleCrossings() gin.HandlerFunc {
	return func(c *gin.Context) {
		seconds := c.Query("seconds")
		since, err := strconv.Atoi(seconds)
		if err != nil {
			_ = c.AbortWithError(500, errors.New("seconds must be a number"))
			return
		}
		crossings, err := s.db.GetCrossings(time.Duration(since) * time.Second)
		if err != nil {
			_ = c.AbortWithError(500, errors.New(fmt.Sprintf("unable to get crossings: %e", err)))
			return
		}
		log.Printf("got %d crossings", len(crossings))
		c.String(200, string(formatCrossings(crossings)))
	}
}

func formatCrossings(crossings []*types.Crossing) []byte {
	fc := geojson.NewFeatureCollection()
	for _, crossing := range crossings {
		pointFeature := geojson.NewPointFeature([]float64{crossing.Longitude, crossing.Latitude})
		pointFeature.Properties = map[string]interface{}{
			"id": fmt.Sprintf("%f %f %d",crossing.Longitude, crossing.Latitude, crossing.Weight),
			"weight": crossing.Weight,
		}
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
