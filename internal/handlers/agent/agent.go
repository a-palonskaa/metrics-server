package agent

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func SendRequest(client *resty.Client, endpoint string, mType string, name string, val fmt.Stringer) error {
	body := Metrics{
		ID:    name,
		MType: mType,
	}

	switch mType {
	case "gauge":
		gVal, _ := val.(memstorage.Gauge)
		fVal := float64(gVal)
		body.Value = &fVal
	case "counter":
		cVal, _ := val.(memstorage.Counter)
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
