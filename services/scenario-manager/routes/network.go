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

//Function for creating a network-config
//@Summary Insert a network-config to database
//@Description Create a network-config
//@Tags network-config
//@Accept json
//@Product json
//@Param network_config body entities.NetworkConfig true "NetworkConfig"
//@Success 200 {object} entities.NetworkConfig "network-config data with success message"
//@Failure 500 {object} nil "network-config null with failure message"
//@Router /api/network-config [post]
func CreateNetwork(c *fiber.Ctx) error {
	var network entities.NetworkConfig

	if err := c.BodyParser(&network); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	var id = utils.GenUUID()
	network.Id = id
	network.Status = entities.STATUS_NONE
	network.CreatedAt = time.Now()
	network.UpdatedAt = time.Now()

	database.Set(utils.KEY_PREFIX_NETWORK+id, &network)

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Network config has been created successfully.", network))
}

//Function for retriving all network-config
//@Summary Get all network-config from database
//@Description Get all network-config
//@Tags network-config
//@Accept json
//@Product json
//@Success 200 {object} []entities.NetworkConfig "array of network-config with success message"
//@Failure 404 {object} nil "null network-config data with error message"
//@Router /api/network-config [get]
func GetNetworks(c *fiber.Ctx) error {
	var values map[string]string

	values, err := database.GetAllValuesWithKeyPrefix(utils.KEY_PREFIX_NETWORK)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	if len(values) < 1 {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", errors.New("network config not present").Error(), nil))
	}

	var responseNetworks []entities.NetworkConfig

	for _, value := range values {
		var network entities.NetworkConfig

		_ = json.Unmarshal([]byte(value), &network)
		responseNetworks = append(responseNetworks, network)
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"status": "OK", "message": "OK", "data": responseNetworks})
}

//Function for retriving a network-config
//@Summary Get a network-config from database
//@Description Get a network-config
//@Tags network-config
//@Accept json
//@Product json
//@Param id path string true "NetworkId"
//@Success 200 {object} entities.NetworkConfig "network-config data with success message"
//@Failure 404 {object} nil "network-config data with null and error message"
//@Router /api/network-config/{id} [get]
func GetNetwork(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Network config id is missing!", nil))
	}

	var network entities.NetworkConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_NETWORK, &network); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Network config not found!", nil))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "OK", network))
}

//Function for updating a network-config
//@Summary Update a network-config to database
//@Description Update a network-config
//@Tags network-config
//@Accept json
//@Product json
//@Param id path string true "NetworkId"
//@Param network_config body string true "NetworkConfig"
//@Success 200 {object} entities.NetworkConfig "network-config data with success message"
//@Failure 500 {object} nil "network-config null with failure message"
//@Router /api/network-config/{id} [put]
func UpdateNetwork(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Network config id is missing!", nil))
	}

	var network entities.NetworkConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_NETWORK, &network); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Network config not found!", nil))
	}

	var updateNetwork entities.NetworkConfig
	if err := c.BodyParser(&updateNetwork); err != nil {
		c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	utils.EntityUpdateCheck(utils.UpdateChecker, &network, &updateNetwork)
	network.UpdatedAt = time.Now()

	database.Set(utils.KEY_PREFIX_NETWORK+id, &network)

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "OK", network))
}

// Function for delete a network-config
// @Summary Delete a network-config from database
// @Description Delete a network-config
// @Tags network-config
// @Accept json
// @Product json
// @Param id path string true "NetworkId"
// @Success 200 {object} entities.NetworkConfig "network-config data with success message"
// @Failure 404 {object} nil "network-config data with null and error message"
// @Router /api/network-config/{id} [delete]
func DeleteNetwork(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Network config id is missing!", nil))
	}

	var network entities.NetworkConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_NETWORK, &network); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Network config not found!", nil))
	}

	if err := database.Del(utils.KEY_PREFIX_NETWORK + id); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Network config has been deleted!", nil))
}
