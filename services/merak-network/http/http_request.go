package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func RequestCall(url, method string, bodyIn interface{}) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	body, _ := json.Marshal(bodyIn)
	log.Printf("body %s", string(body))
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("Got error %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("user-agent", "golang application")
	//req.Header.Add("foo", "bar1")
	//req.Header.Add("foo", "bar2")
	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Got error %s", err.Error())
	}
	defer response.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(response.Body)
	bodyString := string(bodyBytes)
	log.Printf("response: %s", bodyString)
	return bodyString, nil
}
