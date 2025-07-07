package agent

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

func SendRequest(client *resty.Client, endpoint string, mType string, name string, val fmt.Stringer) error {
	_, err := client.SetBaseURL("http://"+endpoint).R().SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"mType": mType,
			"name":  name,
			"val":   val.String(),
		}).Post("/update/{mType}/{name}/{val}")

	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
