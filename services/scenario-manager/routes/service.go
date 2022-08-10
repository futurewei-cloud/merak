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

//Function for creating a service-config
//@Summary Insert a service-config to database
//@Description Create a service-config
//@Tags service-config
//@Accept json
//@Product json
//@Param service_config body entities.ServiceConfig true "ServiceConfig"
//@Success 200 {object} entities.ServiceConfig "service-config data with success message"
//@Failure 500 {object} nil "service-config null with failure message"
//@Router /api/service-config [post]
func CreateService(c *fiber.Ctx) error {
	var service entities.ServiceConfig

	if err := c.BodyParser(&service); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	var id = utils.GenUUID()
	service.Id = id
	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()

	database.Set(utils.KEY_PREFIX_SERVICE+id, &service)

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Service config has been created successfully.", service))
}

//Function for retriving all service-config
//@Summary Get all service-config from database
//@Description Get all service-config
//@Tags service-config
//@Accept json
//@Product json
//@Success 200 {object} []entities.ServiceConfig "array of service-config with success message"
//@Failure 404 {object} nil "null service-config data with error message"
//@Router /api/service-config [get]
func GetServices(c *fiber.Ctx) error {
	var values map[string]string

	values, err := database.GetAllValuesWithKeyPrefix(utils.KEY_PREFIX_SERVICE)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	if len(values) < 1 {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", errors.New("service config not present").Error(), nil))
	}

	var responseNetworks []entities.ServiceConfig

	for _, value := range values {
		var service entities.ServiceConfig

		_ = json.Unmarshal([]byte(value), &service)
		responseNetworks = append(responseNetworks, service)
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"status": "OK", "message": "OK", "data": responseNetworks})
}

//Function for retriving a service-config
//@Summary Get a service-config from database
//@Description Get a service-config
//@Tags service-config
//@Accept json
//@Product json
//@Param id path string true "NetworkId"
//@Success 200 {object} entities.ServiceConfig "service-config data with success message"
//@Failure 404 {object} nil "service-config data with null and error message"
//@Router /api/service-config/{id} [get]
func GetService(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Service config id is missing!", nil))
	}

	var service entities.ServiceConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_SERVICE, &service); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Service config not found!", nil))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "OK", service))
}

//Function for updating a service-config
//@Summary Update a service-config to database
//@Description Update a service-config
//@Tags service-config
//@Accept json
//@Product json
//@Param id path string true "NetworkId"
//@Param Service_config body string true "ServiceConfig"
//@Success 200 {object} entities.ServiceConfig "service-config data with success message"
//@Failure 500 {object} nil "service-config null with failure message"
//@Router /api/service-config/{id} [put]
func UpdateService(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Service config id is missing!", nil))
	}

	var service entities.ServiceConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_SERVICE, &service); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Service config not found!", nil))
	}

	var updateService entities.ServiceConfig
	if err := c.BodyParser(&updateService); err != nil {
		c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	utils.EntityUpdateCheck(utils.UpdateChecker, &service, &updateService)
	service.UpdatedAt = time.Now()

	database.Set(utils.KEY_PREFIX_SERVICE+id, &service)

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "OK", service))
}

// Function for delete a service-config
// @Summary Delete a service-config from database
// @Description Delete a service-config
// @Tags service-config
// @Accept json
// @Product json
// @Param id path string true "NetworkId"
// @Success 200 {object} entities.ServiceConfig "service-config data with success message"
// @Failure 404 {object} nil "service-config data with null and error message"
// @Router /api/service-config/{id} [delete]
func DeleteService(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Service config id is missing!", nil))
	}

	var service entities.ServiceConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_SERVICE, &service); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Service config not found!", nil))
	}

	if err := database.Del(utils.KEY_PREFIX_SERVICE + id); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Service config has been deleted!", nil))
}
