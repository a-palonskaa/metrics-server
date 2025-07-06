package agent

import (
	"fmt"
	"github.com/go-resty/resty/v2"

	st "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func SendRequest(client *resty.Client, endpoint string, kind string, name string, val st.Stringer) error {
	_, err := client.R().SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"endpointAddr": endpoint,
			"kind":         kind,
			"name":         name,
			"val":          val.String(),
		}).Post("http://{endpointAddr}/update/{kind}/{name}/{val}")

	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
