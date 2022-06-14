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
//@Router /v1/topologies [get]
func GetTopologies(c *fiber.Ctx) error {
	var values map[string]string

	values, err := database.GetAllValuesWithKeyPrefix(utils.KEY_PREFIX_TOPOLOGY)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	if len(values) < 1 {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", errors.New("Topology not present!!!").Error(), nil))
	}

	var responseTopologies []entities.TopologyConfig

	for _, value := range values {
		var topology entities.TopologyConfig

		err = json.Unmarshal([]byte(value), &topology)
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
//@Router /v1/topologies/{id} [get]
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
//@Router /v1/topologies/{id} [put]
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

	topology.Name = updateTopology.Name
	topology.NumberOfVhosts = updateTopology.NumberOfVhosts
	topology.NumberOfRacks = updateTopology.NumberOfRacks
	topology.TopoType = updateTopology.TopoType
	topology.DataPlaneCidr = updateTopology.DataPlaneCidr
	topology.NumberOfGateways = updateTopology.NumberOfGateways
	topology.GatewayIPs = updateTopology.GatewayIPs
	topology.VNodes = updateTopology.VNodes
	topology.VLinks = updateTopology.VLinks
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
// @Router /v1/topologies/{id} [delete]
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
