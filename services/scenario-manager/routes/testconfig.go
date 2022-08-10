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

package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/futurewei-cloud/merak/services/scenario-manager/database"
	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
	"github.com/futurewei-cloud/merak/services/scenario-manager/utils"
	"github.com/gofiber/fiber/v2"
)

//Function for creating a test-config
//@Summary Insert a test-config to database
//@Description Create a test-config
//@Tags test-config
//@Accept json
//@Product json
//@Param test_config body entities.TestConfig true "TestConfig"
//@Success 200 {object} entities.TestConfig "Compute data with success message"
//@Failure 500 {object} nil "Compute null with failure message"
//@Router /api/test-config [post]
func CreateTestConfig(c *fiber.Ctx) error {
	var test entities.TestConfig

	if err := c.BodyParser(&test); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	var id = utils.GenUUID()
	test.Id = id
	test.Status = entities.STATUS_NONE
	test.CreatedAt = time.Now()
	test.UpdatedAt = time.Now()

	database.Set(utils.KEY_PREFIX_TEST+id, &test)

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Test config has been created successfully.", test))
}

//Function for retriving all test-config
//@Summary Get all test-config from database
//@Description Get all test-config
//@Tags test-config
//@Accept json
//@Product json
//@Success 200 {object} []entities.TestConfig "array of test-config with success message"
//@Failure 404 {object} nil "null test-config data with error message"
//@Router /api/test-config [get]
func GetTestConfigs(c *fiber.Ctx) error {
	var values map[string]string

	values, err := database.GetAllValuesWithKeyPrefix(utils.KEY_PREFIX_TEST)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	if len(values) < 1 {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", errors.New("test config not present").Error(), nil))
	}

	var responseTests []entities.TestConfig

	for _, value := range values {
		var test entities.TestConfig

		_ = json.Unmarshal([]byte(value), &test)
		responseTests = append(responseTests, test)
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"status": "OK", "message": "OK", "data": responseTests})
}

//Function for retriving a test-config
//@Summary Get a test-config from database
//@Description Get a test-config
//@Tags test-config
//@Accept json
//@Product json
//@Param id path string true "ComputeConfId"
//@Success 200 {object} entities.TestConfig "test-config data with success message"
//@Failure 404 {object} nil "test-config data with null and error message"
//@Router /api/test-config/{id} [get]
func GetTestConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Test config id is missing!", nil))
	}

	var test entities.TestConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_TEST, &test); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Test config not found!", nil))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "OK", test))
}

//Function for updating a test-config
//@Summary Update a test-config to database
//@Description Update a test-config
//@Tags test-config
//@Accept json
//@Product json
//@Param id path string true "ComputeConfId"
//@Param compute_config body string true "TestConfig"
//@Success 200 {object} entities.TestConfig "compute_config data with success message"
//@Failure 500 {object} nil "compute_config null with failure message"
//@Router /api/test-config/{id} [put]
func UpdateTestConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Test config id is missing!", nil))
	}

	var test entities.TestConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_TEST, &test); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Test config not found!", nil))
	}

	var updateTest entities.TestConfig
	if err := c.BodyParser(&updateTest); err != nil {
		c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	utils.EntityUpdateCheck(utils.UpdateChecker, &test, &updateTest)
	test.UpdatedAt = time.Now()

	database.Set(utils.KEY_PREFIX_TEST+id, &test)

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "OK", test))
}

// Function for delete a test-config
// @Summary Delete a test-config from database
// @Description Delete a test-config
// @Tags test-config
// @Accept json
// @Product json
// @Param id path string true "ComputeConfId"
// @Success 200 {object} entities.TestConfig "test-config data with success message"
// @Failure 404 {object} nil "test-config data with null and error message"
// @Router /api/test-config/{id} [delete]
func DeleteTestConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Test config id is missing!", nil))
	}

	var test entities.TestConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_TEST, &test); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Test config not found!", nil))
	}

	if err := database.Del(utils.KEY_PREFIX_TEST + id); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Test config has been deleted!", nil))
}
