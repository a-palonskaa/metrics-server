package agent

import (
	"fmt"
	"io"
	"net/http"
)

func SendRequest(client *http.Client, endpoint string, kind string, name string, val interface{}) error {
	url := fmt.Sprintf("http://%s/update/%s/%s/%v", endpoint, kind, name, val)
	response, err := client.Post(url, "text/html", nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
