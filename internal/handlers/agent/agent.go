package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func SendRequest(client *http.Client, endpoint string, kind string, name string, val interface{}) error {
	url := fmt.Sprintf("http://%s/update/%s/%s/%v", endpoint, kind, name, val)
	response, err := client.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if _, err = io.Copy(io.Discard, response.Body); err != nil {
		fmt.Println(err)
		return err
	}

	if err := response.Body.Close(); err != nil {
		log.Printf("failed to lcose response body: %s", err)
	}
	return nil
}
