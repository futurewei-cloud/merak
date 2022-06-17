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
			body: map[string]string{"id": "", "name": "testScenario1", "project_id": "1", "topology_id": "1",
				"service_config_id": "1", "network_config_id": "1", "compute_config_id": "1", "test_config_id": "1"},
			expectedError: false,
			expectedCode:  200,
			expectedBody:  `{"status": "OK", "message": "Scenario Has been created successfully.", "data": {"id": "", "name": "testScenario1", "project_id": "1", "topology_id": "1","service_config_id": "1", "network_config_id": "1", "compute_config_id": "1", "test_config_id": "1"}}`,
		},
		{
			description:   "non existing route",
			route:         "/i-dont-exist",
			expectedError: false,
			expectedCode:  404,
			expectedBody:  "Cannot GET /i-dont-exist",
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
	sa.ScenarioId = "194d15c6aa3b4820bb640af26a22f2bf"
	var ssa entities.ServiceAction
	ssa.ServiceName = "topology"
	ssa.Action = "deploy"
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

func TestIndexRoute(t *testing.T) {
	// Define a structure for specifying input and output
	// data of a single test case. This structure is then used
	// to create a so called test map, which contains all test
	// cases, that should be run for testing this function
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
			description:   "index route",
			route:         "/",
			expectedError: false,
			expectedCode:  200,
			expectedBody:  "OK",
		},
		{
			description:   "non existing route",
			route:         "/i-dont-exist",
			expectedError: false,
			expectedCode:  404,
			expectedBody:  "Cannot GET /i-dont-exist",
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
		body, err := ioutil.ReadAll(res.Body)

		// Reading the response body should work everytime, such that
		// the err variable should be nil
		utils.AssertEqual(t, nil, err, test.description)

		// Verify, that the reponse body equals the expected body
		utils.AssertEqual(t, test.expectedBody, string(body), test.description)
	}
}

func TestGetCompute(t *testing.T) {
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
			description:   "get compute-config",
			route:         "/api/compute-config/489975ac36dc49a99054be828b3dc542",
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
		body, err := ioutil.ReadAll(res.Body)

		// Reading the response body should work everytime, such that
		// the err variable should be nil
		utils.AssertEqual(t, nil, err, test.description)

		// Verify, that the reponse body equals the expected body
		utils.AssertEqual(t, test.expectedBody, string(body), test.description)
	}
}
