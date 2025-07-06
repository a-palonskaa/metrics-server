package agent

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	//	"io"
	//	"log"

	st "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func SendRequest(client *resty.Client, endpoint string, kind string, name string, val st.Stringer) error {
	//url := fmt.Sprintf("http://%s/update/%s/%s/%v", endpoint, kind, name, val)
	//response, err := client.Post(url, "text/plain", nil)
	//if err != nil {
	//	fmt.Println(err)
	//	return err
	//}

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

	//if _, err = io.Copy(io.Discard, response.Body); err != nil {
	//	fmt.Println(err)
	//	return err
	//}
	//
	//if err := response.Body.Close(); err != nil {
	//	log.Printf("failed to lcose response body: %s", err)
	//}
	return nil
}
