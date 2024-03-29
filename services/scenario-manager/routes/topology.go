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

//Function for creating a topology
//@Summary Insert a topology to database
//@Description Create a topology
//@Tags topology
//@Accept json
//@Product json
//@Param topology_config body entities.TopologyConfig true "TopologyConfig"
//@Success 200 {object} entities.TopologyConfig "topology data with success message"
//@Failure 500 {object} nil "topology null with failure message"
//@Router /api/topologies [post]
func CreateTopology(c *fiber.Ctx) error {
	var topology entities.TopologyConfig

	if err := c.BodyParser(&topology); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	var id = utils.GenUUID()
	topology.Id = id
	topology.Status = entities.STATUS_NONE
	topology.CreatedAt = time.Now()
	topology.UpdatedAt = time.Now()

	database.Set(utils.KEY_PREFIX_TOPOLOGY+id, &topology)

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Topology Has been created successfully.", topology))
}

//Function for retriving all topologies
//@Summary Get all topologies from database
//@Description Get all topologies
//@Tags topology
//@Accept json
//@Product json
//@Success 200 {object} []entities.TopologyConfig "array of topology with success message"
//@Failure 404 {object} nil "null topology data with error message"
//@Router /api/topologies [get]
func GetTopologies(c *fiber.Ctx) error {
	var values map[string]string

	values, err := database.GetAllValuesWithKeyPrefix(utils.KEY_PREFIX_TOPOLOGY)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	if len(values) < 1 {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", errors.New("topology not present").Error(), nil))
	}

	var responseTopologies []entities.TopologyConfig

	for _, value := range values {
		var topology entities.TopologyConfig

		_ = json.Unmarshal([]byte(value), &topology)
		responseTopologies = append(responseTopologies, topology)
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"status": "OK", "message": "OK", "data": responseTopologies})
}

//Function for retriving a topology
//@Summary Get a topology from database
//@Description Get a topology
//@Tags topology
//@Accept json
//@Product json
//@Param id path string true "TopologyId"
//@Success 200 {object} entities.TopologyConfig "topology data with success message"
//@Failure 404 {object} nil "topology data with null and error message"
//@Router /api/topologies/{id} [get]
func GetTopology(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Topology id is missing!", nil))
	}

	var topology entities.TopologyConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_TOPOLOGY, &topology); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Topology not found!", nil))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "OK", topology))
}

//Function for updating a topology
//@Summary Update a topology to database
//@Description Update a topology
//@Tags topology
//@Accept json
//@Product json
//@Param id path string true "TopologyId"
//@Param topology_config body string true "TopologyConfig"
//@Success 200 {object} entities.TopologyConfig "topology data with success message"
//@Failure 500 {object} nil "topology null with failure message"
//@Router /api/topologies/{id} [put]
func UpdateTopology(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Topology id is missing!", nil))
	}

	var topology entities.TopologyConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_TOPOLOGY, &topology); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Topology not found!", nil))
	}

	var updateTopology entities.TopologyConfig
	if err := c.BodyParser(&updateTopology); err != nil {
		c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	utils.EntityUpdateCheck(utils.UpdateChecker, &topology, &updateTopology)
	topology.UpdatedAt = time.Now()

	database.Set(utils.KEY_PREFIX_TOPOLOGY+id, &topology)

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "OK", topology))
}

// Function for delete a topology
// @Summary Delete a topology from database
// @Description Delete a topology
// @Tags topology
// @Accept json
// @Product json
// @Param id path string true "TopologyId"
// @Success 200 {object} entities.TopologyConfig "topology data with success message"
// @Failure 404 {object} nil "topology data with null and error message"
// @Router /api/topologies/{id} [delete]
func DeleteTopology(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Topology id is missing!", nil))
	}

	var topology entities.TopologyConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_TOPOLOGY, &topology); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Topology not found!", nil))
	}

	if err := database.Del(utils.KEY_PREFIX_TOPOLOGY + id); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Topology has been deleted!", nil))
}
