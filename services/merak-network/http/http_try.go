//package http
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

//var (
//	vpc_endpoint = "54.188.252.43"
//	vpc_port     = "30001"
//	project_id   = "123456789"
//)

func main() {
	call("http://54.188.252.43:30001/project/123456789/vpcs", "GET")
}
func call(url, method string) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}
	//req.Header.Set("user-agent", "golang application")
	//req.Header.Add("foo", "bar1")
	//req.Header.Add("foo", "bar2")
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}
	defer response.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(response.Body)
	bodyString := string(bodyBytes)
	log.Printf("response: %s", bodyString)
	return nil
}

// post
//func main() {
//
//	data := url.Values{
//		"admin_state_up":        {"True"},
//		"revision_number":       {"0"},
//		"cidr":                  {"10.9.0.0/16"},
//		"default":               {"True"},
//		"description":           {"vpc"},
//		"dns_domain":            {"domain"},
//		"id":                    {"9192a4d4-ffff-4ece-b3f0-8d36e3d88009"},
//		"is_default":            {"True"},
//		"mtu":                   {"1400"},
//		"name":                  {"sample_vpc"},
//		"port_security_enabled": {"True"},
//		"project_id":            {"123456789"},
//	}
//	vpc_url := "http://" + vpc_endpoint + ":" + vpc_port + "/project/" + project_id + "/vpcs"
//
//	resp, err := http.PostForm(vpc_url, data)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	var res map[string]interface{}
//
//	json.NewDecoder(resp.Body).Decode(&res)
//
//	fmt.Println(resp)
//}

// get
//func main() {
//	fmt.Println("1. Performing Http Get...")
//	//resp, err := http.Get("https://jsonplaceholder.typicode.com/todos/1")
//	resp, err := http.Get("http://54.188.252.43:30001/project/123456789/vpcs")
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	defer resp.Body.Close()
//	bodyBytes, _ := ioutil.ReadAll(resp.Body)
//
//	// Convert response body to string
//	bodyString := string(bodyBytes)
//	fmt.Println("API Response as String:\n" + bodyString)
//
//	// Convert response body to Todo struct
//	//var todoStruct Todo
//	//json.Unmarshal(bodyBytes, &todoStruct)
//	//fmt.Printf("API Response as struct %+v\n", todoStruct)
//}
