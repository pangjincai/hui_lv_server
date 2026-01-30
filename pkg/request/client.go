package request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var Client = &http.Client{
	Timeout: 10 * time.Second,
}

func Get(url string, target interface{}) error {
	resp, err := Client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, target)
}

func GetRaw(url string) ([]byte, error) {
	resp, err := Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
