package agent

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	errhandler "github.com/a-palonskaa/metrics-server/internal/err_handlers"
	metrics "github.com/a-palonskaa/metrics-server/internal/metrics"
	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func SendRequest(client *resty.Client, endpoint string, body metrics.MetricsS) error {
	if len(body) == 0 {
		return nil
	}

	jsonData, err := body.MarshalJSON()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jsonData); err != nil {
		log.Error().Err(err)
		return err
	}
	if err := gz.Close(); err != nil {
		log.Error().Err(err)
		return err
	}

	_, err = client.SetBaseURL("http://"+endpoint).R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf.Bytes()).
		Post("/updates/")
	if err != nil {
		log.Error().Err(err).Msg("failed to send request")
		return err
	}
	log.Info().Msgf("sent metrics to server, %s", "http://"+endpoint+"/updates/")
	return nil
}

func MakeSendMetricsFunc(client *resty.Client, endpointAddr string) func() {
	return func() {
		var metric metrics.Metrics
		err := errhandler.RetriableErrHadler(
			func() error {
				var body []metrics.Metrics
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

				return SendRequest(client, endpointAddr, metrics.MetricsS(body))
			}, errhandler.CompareErrAgent)
		if err != nil {
			log.Error().Err(err).Msg("error sending metrics")
		}
	}
}
