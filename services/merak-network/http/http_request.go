package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func RequestCall(url, method string, bodyIn interface{}, headers []string) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 20,
	}
	body, _ := json.Marshal(bodyIn)
	log.Printf("body %s", string(body))
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("Got error %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	if headers != nil {
		for _, header := range headers {
			req.Header.Add(strings.Split(header, " ")[0], strings.Split(header, " ")[1])
		}
	}
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
	//log.Printf("response: %s", bodyString)
	if response.StatusCode < 200 || response.StatusCode > 299 {
		//log.Printf("RequestCall Fail Code %s", response.StatusCode)
		//log.Printf("RequestCall Fail %s", bodyString)
		log.Println("RequestCall Fail ", response.StatusCode, bodyString)
		return "", fmt.Errorf(bodyString)
	}
	return bodyString, nil
}
