package utils

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

func Send(host string, port int, cmd string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	url := "http://" + host + ":" + strconv.Itoa(port) + "/api/v1/machine/create"
	contentType := "application/json"
	req, err := http.NewRequest("POST", url, strings.NewReader(cmd))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}
