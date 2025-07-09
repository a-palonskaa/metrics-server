package agent

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/metrics"
)

func SendRequest(client *resty.Client, endpoint string, mType string, name string, val fmt.Stringer) error {
	body := metrics.Metrics{
		ID:    name,
		MType: mType,
	}

	switch mType {
	case "gauge":
		gVal, _ := val.(metrics.Gauge)
		fVal := float64(gVal)
		body.Value = &fVal
	case "counter":
		cVal, _ := val.(metrics.Counter)
		iVal := int64(cVal)
		body.Delta = &iVal
	default:
		log.Error().Msg("unknown type")
		return fmt.Errorf("unknown type %s", mType)
	}

	_, err := client.SetBaseURL("http://"+endpoint).R().
		SetHeader("Content-Type", "application/json").SetBody(body).
		Post("/update/")

	if err != nil {
		log.Error().Err(err)
		return err
	}
	return nil
}
