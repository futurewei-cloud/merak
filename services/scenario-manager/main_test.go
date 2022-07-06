package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
	utils "github.com/gofiber/fiber/v2/utils"
)

func TestCreateScenarioAndOthers(t *testing.T) {
	// Define a structure for specifying input and output
	// data of a single test case. This structure is then used
	// to create a so called test map, which contains all test
	// cases, that should be run for testing this function
	var topology_id string
	var netconfig_id string
	var compute_id string
	var service_id string
	var test_id string

	tests := []struct {
		description string

		// Test input
		route string
		body  map[string]string

		// Expected output
		expectedError bool
		expectedCode  int
		expectedBody  string
	}{
		{
			description:   "create a topology",
			route:         "/api/topologies",
			body:          map[string]string{"name": "topology-test1"},
			expectedError: false,
			expectedCode:  200,
			expectedBody:  `{"status": "OK", "message": "Scenario Has been created successfully.", "data": {"id": "", "name": "testScenario1", "project_id": "1", "topology_id": "1","service_config_id": "1", "network_config_id": "1", "compute_config_id": "1", "test_config_id": "1"}}`,
		},
		{
			description:   "create a network",
			route:         "/api/network-config",
			body:          map[string]string{"name": "network-test1"},
			expectedError: false,
			expectedCode:  200,
			expectedBody:  `{"status": "OK", "message": "Scenario Has been created successfully.", "data": {"id": "", "name": "testScenario1", "project_id": "1", "topology_id": "1","service_config_id": "1", "network_config_id": "1", "compute_config_id": "1", "test_config_id": "1"}}`,
		},
		{
			description:   "create a service",
			route:         "/api/service-config",
			body:          map[string]string{"name": "service-test1"},
			expectedError: false,
			expectedCode:  200,
			expectedBody:  `{"status": "OK", "message": "Scenario Has been created successfully.", "data": {"id": "", "name": "testScenario1", "project_id": "1", "topology_id": "1","service_config_id": "1", "network_config_id": "1", "compute_config_id": "1", "test_config_id": "1"}}`,
		},
		{
			description:   "create a compute",
			route:         "/api/compute-config",
			body:          map[string]string{"name": "compute-test1"},
			expectedError: false,
			expectedCode:  200,
			expectedBody:  `{"status": "OK", "message": "Scenario Has been created successfully.", "data": {"id": "", "name": "testScenario1", "project_id": "1", "topology_id": "1","service_config_id": "1", "network_config_id": "1", "compute_config_id": "1", "test_config_id": "1"}}`,
		},
		{
			description:   "create a test",
			route:         "/api/test-config",
			body:          map[string]string{"name": "test-test1"},
			expectedError: false,
			expectedCode:  200,
			expectedBody:  `{"status": "OK", "message": "Scenario Has been created successfully.", "data": {"id": "", "name": "testScenario1", "project_id": "1", "topology_id": "1","service_config_id": "1", "network_config_id": "1", "compute_config_id": "1", "test_config_id": "1"}}`,
		},
		{
			description: "create a scenario",
			route:       "/api/scenarios",
			body: map[string]string{"name": "testScenario1", "project_id": "1", "topology_id": topology_id,
				"service_config_id": service_id, "network_config_id": netconfig_id, "compute_config_id": compute_id, "test_config_id": test_id},
			expectedError: false,
			expectedCode:  200,
			expectedBody:  `{"status": "OK", "message": "Scenario Has been created successfully.", "data": {"id": "", "name": "testScenario1", "project_id": "1", "topology_id": "1","service_config_id": "1", "network_config_id": "1", "compute_config_id": "1", "test_config_id": "1"}}`,
		},
	}

	// Setup the app as it is done in the main function
	app := Setup()

	// Iterate through test single test cases
	for _, test := range tests {
		// Create a new http request with the route
		// from the test case

		if test.description == "create a scenario" {
			test.body["topology_id"] = topology_id
			test.body["service_config_id"] = service_id
			test.body["network_config_id"] = netconfig_id
			test.body["compute_config_id"] = compute_id
			test.body["test_config_id"] = test_id
		}

		reqbody, _ := json.Marshal(test.body)
		req, _ := http.NewRequest(
			"POST",
			test.route,
			bytes.NewReader(reqbody),
		)
		req.Header.Set("Content-Type", "application/json")

		// Perform the request plain with the app.
		// The -1 disables request latency.
		res, err := app.Test(req, -1)

		// verify that no error occured, that is not expected
		utils.AssertEqual(t, test.expectedError, err != nil, test.description)

		// As expected errors lead to broken responses, the next
		// test case needs to be processed
		if test.expectedError {
			continue
		}

		// Verify if the status code is as expected
		utils.AssertEqual(t, test.expectedCode, res.StatusCode, test.description)

		// Read the response body
		body, err := ioutil.ReadAll(res.Body)

		// Reading the response body should work everytime, such that
		// the err variable should be nil
		utils.AssertEqual(t, nil, err, test.description)

		var entity map[string]interface{}

		switch test.description {
		case "create a topology":
			err := json.Unmarshal([]byte(body), &entity)
			utils.AssertEqual(t, nil, err, test.description)
			data := entity["data"].(map[string]interface{})
			topology_id = data["id"].(string)
		case "create a network":
			err = json.Unmarshal([]byte(body), &entity)
			utils.AssertEqual(t, nil, err, test.description)
			data := entity["data"].(map[string]interface{})
			netconfig_id = data["id"].(string)
		case "create a service":
			err = json.Unmarshal([]byte(body), &entity)
			utils.AssertEqual(t, nil, err, test.description)
			data := entity["data"].(map[string]interface{})
			service_id = data["id"].(string)
		case "create a compute":
			err = json.Unmarshal([]byte(body), &entity)
			utils.AssertEqual(t, nil, err, test.description)
			data := entity["data"].(map[string]interface{})
			compute_id = data["id"].(string)
		case "create a test":
			err = json.Unmarshal([]byte(body), &entity)
			utils.AssertEqual(t, nil, err, test.description)
			data := entity["data"].(map[string]interface{})
			test_id = data["id"].(string)
		}
		// Verify, that the reponse body equals the expected body
		//utils.AssertEqual(t, test.expectedBody, string(body), test.description)
	}
}

func TestCreateScenario(t *testing.T) {
	// Define a structure for specifying input and output
	// data of a single test case. This structure is then used
	// to create a so called test map, which contains all test
	// cases, that should be run for testing this function
	tests := []struct {
		description string

		// Test input
		route string
		body  map[string]string

		// Expected output
		expectedError bool
		expectedCode  int
		expectedBody  string
	}{
		{
			description: "create a scenario",
			route:       "/api/scenarios",
			body: map[string]string{"name": "testScenario1", "project_id": "1", "topology_id": "1",
				"service_config_id": "1", "network_config_id": "1", "compute_config_id": "1", "test_config_id": "1"},
			expectedError: false,
			expectedCode:  200,
			expectedBody:  `{"status": "OK", "message": "Scenario Has been created successfully.", "data": {"id": "", "name": "testScenario1", "project_id": "1", "topology_id": "1","service_config_id": "1", "network_config_id": "1", "compute_config_id": "1", "test_config_id": "1"}}`,
		},
	}

	// Setup the app as it is done in the main function
	app := Setup()

	// Iterate through test single test cases
	for _, test := range tests {
		// Create a new http request with the route
		// from the test case
		reqbody, _ := json.Marshal(test.body)
		req, _ := http.NewRequest(
			"POST",
			test.route,
			bytes.NewReader(reqbody),
		)
		req.Header.Set("Content-Type", "application/json")

		// Perform the request plain with the app.
		// The -1 disables request latency.
		res, err := app.Test(req, -1)

		// verify that no error occured, that is not expected
		utils.AssertEqual(t, test.expectedError, err != nil, test.description)

		// As expected errors lead to broken responses, the next
		// test case needs to be processed
		if test.expectedError {
			continue
		}

		// Verify if the status code is as expected
		utils.AssertEqual(t, test.expectedCode, res.StatusCode, test.description)

		// Read the response body
		//body, err := ioutil.ReadAll(res.Body)

		// Reading the response body should work everytime, such that
		// the err variable should be nil
		utils.AssertEqual(t, nil, err, test.description)

		// Verify, that the reponse body equals the expected body
		//utils.AssertEqual(t, test.expectedBody, string(body), test.description)
	}
}

func TestScenarioActions(t *testing.T) {
	var sa entities.ScenarioAction
	sa.ScenarioId = "d6f044df409d4836930ee88b540b2610"
	var ssa entities.ServiceAction
	ssa.ServiceName = "topology"
	ssa.Action = "DEPLOY"
	sa.Services = append(sa.Services, ssa)

	// Setup the app as it is done in the main function
	app := Setup()

	reqbody, _ := json.Marshal(sa)
	req, _ := http.NewRequest(
		"POST",
		"/api/scenarios/actions",
		bytes.NewReader(reqbody),
	)
	req.Header.Set("Content-Type", "application/json")

	// Perform the request plain with the app.
	// The -1 disables request latency.
	res, err := app.Test(req, -1)

	// verify that no error occured, that is not expected
	utils.AssertEqual(t, false, err != nil, "deploy a scenario")

	// Verify if the status code is as expected
	utils.AssertEqual(t, 200, res.StatusCode, "deploy a scenario")

	// Read the response body
	//body, err := ioutil.ReadAll(res.Body)

	// Reading the response body should work everytime, such that
	// the err variable should be nil
	// utils.AssertEqual(t, nil, err, test.description)

	// Verify, that the reponse body equals the expected body
	//utils.AssertEqual(t, test.expectedBody, string(body), test.description)
}

func TestSAGetTopology(t *testing.T) {
	var sa entities.ScenarioAction
	sa.ScenarioId = "d6f044df409d4836930ee88b540b2610"
	var ssa entities.ServiceAction
	ssa.ServiceName = "topology"
	ssa.Action = "CHECK"
	sa.Services = append(sa.Services, ssa)

	// Setup the app as it is done in the main function
	app := Setup()

	reqbody, _ := json.Marshal(sa)
	req, _ := http.NewRequest(
		"POST",
		"/api/scenarios/actions",
		bytes.NewReader(reqbody),
	)
	req.Header.Set("Content-Type", "application/json")

	// Perform the request plain with the app.
	// The -1 disables request latency.
	res, err := app.Test(req, -1)

	// verify that no error occured, that is not expected
	utils.AssertEqual(t, false, err != nil, "deploy a scenario")

	// Verify if the status code is as expected
	utils.AssertEqual(t, 200, res.StatusCode, "deploy a scenario")

	// Read the response body
	//body, err := ioutil.ReadAll(res.Body)

	// Reading the response body should work everytime, such that
	// the err variable should be nil
	// utils.AssertEqual(t, nil, err, test.description)

	// Verify, that the reponse body equals the expected body
	//utils.AssertEqual(t, test.expectedBody, string(body), test.description)
}

func TestGetScenarios(t *testing.T) {
	tests := []struct {
		description string

		// Test input
		route string

		// Expected output
		expectedError bool
		expectedCode  int
		expectedBody  string
	}{
		{
			description:   "get scenarios",
			route:         "/api/scenarios",
			expectedError: false,
			expectedCode:  200,
			expectedBody:  "OK",
		},
	}

	// Setup the app as it is done in the main function

	app := Setup()

	// Iterate through test single test cases
	for _, test := range tests {
		// Create a new http request with the route
		// from the test case
		req, _ := http.NewRequest(
			"GET",
			test.route,
			nil,
		)

		// Perform the request plain with the app.
		// The -1 disables request latency.
		res, err := app.Test(req, -1)

		// verify that no error occured, that is not expected
		utils.AssertEqual(t, test.expectedError, err != nil, test.description)

		// As expected errors lead to broken responses, the next
		// test case needs to be processed
		if test.expectedError {
			continue
		}

		// Verify if the status code is as expected
		utils.AssertEqual(t, test.expectedCode, res.StatusCode, test.description)

		// Read the response body
		//body, err := ioutil.ReadAll(res.Body)

		// Reading the response body should work everytime, such that
		// the err variable should be nil
		utils.AssertEqual(t, nil, err, test.description)

		// Verify, that the reponse body equals the expected body
		//utils.AssertEqual(t, test.expectedBody, string(body), test.description)
	}
}

func TestPutOperations(t *testing.T) {
	tests := []struct {
		description string

		// Test input
		route string
		body  map[string]interface{}

		// Expected output
		expectedError bool
		expectedCode  int
		expectedBody  string
	}{
		// {
		// 	description:   "put a scenario",
		// 	route:         "/api/scenarios/d6f044df409d4836930ee88b540b2610",
		// 	body:          map[string]string{"name": "scenario-test-2", "status": "NONE"},
		// 	expectedError: false,
		// 	expectedCode:  200,
		// 	expectedBody:  "OK",
		// },
		{
			description: "put a topology",
			route:       "/api/topologies/25203a13695a488bb4441bacb1251f2c",
			body: map[string]interface{}{
				"name":             "topology-test-2",
				"status":           "NONE",
				"number_of_vhosts": 10,
				"number_of_racks":  2,
				"vhosts_per_rack":  5,
				"data_plane_cidr":  "10.200.0.0/16",
				"vnodes":           []interface{}{map[string]interface{}{"name": "p1", "type": "vhost", "nics": []interface{}{map[string]interface{}{"name": "eth0", "ip": "10.0.0.1"}}}},
				"vlinks":           []interface{}{map[string]interface{}{"name": "v1", "from": "p1", "to": "p2"}}},
			expectedError: false,
			expectedCode:  200,
			expectedBody:  "OK",
		},
	}

	// Setup the app as it is done in the main function

	app := Setup()

	// Iterate through test single test cases
	for _, test := range tests {
		// Create a new http request with the route
		// from the test case
		reqbody, _ := json.Marshal(test.body)
		req, _ := http.NewRequest(
			"PUT",
			test.route,
			bytes.NewReader(reqbody),
		)
		req.Header.Set("Content-Type", "application/json")

		// Perform the request plain with the app.
		// The -1 disables request latency.
		res, err := app.Test(req, -1)

		// verify that no error occured, that is not expected
		utils.AssertEqual(t, test.expectedError, err != nil, test.description)

		// As expected errors lead to broken responses, the next
		// test case needs to be processed
		if test.expectedError {
			continue
		}

		// Verify if the status code is as expected
		utils.AssertEqual(t, test.expectedCode, res.StatusCode, test.description)

		// Read the response body
		//body, err := ioutil.ReadAll(res.Body)

		// Reading the response body should work everytime, such that
		// the err variable should be nil
		utils.AssertEqual(t, nil, err, test.description)

		// Verify, that the reponse body equals the expected body
		//utils.AssertEqual(t, test.expectedBody, string(body), test.description)
	}
}
