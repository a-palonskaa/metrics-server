package agent

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/metrics"
	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func SendRequest(client *resty.Client, endpoint string, body metrics.MetricsS) error {
	jsonData, err := body.MarshalJSON()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jsonData); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	_, err = client.SetBaseURL("http://"+endpoint).R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf).
		Post("/updates/")
	if err != nil {
		log.Error().Err(err).Msg("failed to send request")
		return err
	}
	return nil
}

func MakeSendMetricsFunc(client *resty.Client, endpointAddr string, backoffScedule []time.Duration) func() {
	return func() {
		var metric metrics.Metrics
		var body []metrics.Metrics
		for _, backoff := range backoffScedule {
			memstorage.MS.Iterate(func(key string, mType string, val fmt.Stringer) {
				metric.ID = key
				metric.MType = mType
				switch mType {
				case "gauge":
					gVal, _ := val.(metrics.Gauge)
					fVal := float64(gVal)
					metric.Value = &fVal
				case "counter":
					cVal, _ := val.(metrics.Counter)
					iVal := int64(cVal)
					metric.Delta = &iVal
				default:
					log.Error().Msg("unknown type")
					return
				}
				body = append(body, metric)
			})

			if err := SendRequest(client, endpointAddr, metrics.MetricsS(body)); err != nil {
				log.Error().Err(err)
			}
			time.Sleep(backoff)
		}

	}
}
