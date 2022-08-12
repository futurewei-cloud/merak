/*
MIT License
Copyright(c) 2022 Futurewei Cloud
    Permission is hereby granted,
    free of charge, to any person obtaining a copy of this software and associated documentation files(the "Software"), to deal in the Software without restriction,
    including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and / or sell copies of the Software, and to permit persons
    to whom the Software is furnished to do so, subject to the following conditions:
    The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
    WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

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
	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Got error %s", err.Error())
	}
	defer response.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(response.Body)
	bodyString := string(bodyBytes)
	if response.StatusCode < 200 || response.StatusCode > 299 {
		log.Println("RequestCall Fail ", response.StatusCode, bodyString)
		return "", fmt.Errorf(bodyString)
	}
	return bodyString, nil
}
