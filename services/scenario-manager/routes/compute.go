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

//Function for creating a compute-config
//@Summary Insert a compute-config to database
//@Description Create a compute-config
//@Tags compute-config
//@Accept json
//@Product json
//@Param compute_config body entities.ComputeConfig true "ComputeConfig"
//@Success 200 {object} entities.ComputeConfig "Compute data with success message"
//@Failure 500 {object} nil "Compute null with failure message"
//@Router /api/compute-config [post]
func CreateCompute(c *fiber.Ctx) error {
	var service entities.ComputeConfig

	if err := c.BodyParser(&service); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	var id = utils.GenUUID()
	service.Id = id
	service.Status = entities.STATUS_NONE
	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()

	database.Set(utils.KEY_PREFIX_COMPUTE+id, &service)

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Compute config has been created successfully.", service))
}

//Function for retriving all compute-config
//@Summary Get all compute-config from database
//@Description Get all compute-config
//@Tags compute-config
//@Accept json
//@Product json
//@Success 200 {object} []entities.ComputeConfig "array of compute-config with success message"
//@Failure 404 {object} nil "null compute-config data with error message"
//@Router /api/compute-config [get]
func GetComputes(c *fiber.Ctx) error {
	var values map[string]string

	values, err := database.GetAllValuesWithKeyPrefix(utils.KEY_PREFIX_COMPUTE)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	if len(values) < 1 {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", errors.New("compute config not present").Error(), nil))
	}

	var responseComputes []entities.ComputeConfig

	for _, value := range values {
		var compute entities.ComputeConfig

		_ = json.Unmarshal([]byte(value), &compute)
		responseComputes = append(responseComputes, compute)
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"status": "OK", "message": "OK", "data": responseComputes})
}

//Function for retriving a compute-config
//@Summary Get a compute-config from database
//@Description Get a compute-config
//@Tags compute-config
//@Accept json
//@Product json
//@Param id path string true "ComputeConfId"
//@Success 200 {object} entities.ComputeConfig "compute-config data with success message"
//@Failure 404 {object} nil "compute-config data with null and error message"
//@Router /api/compute-config/{id} [get]
func GetCompute(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Compute config id is missing!", nil))
	}

	var compute entities.ComputeConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_COMPUTE, &compute); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Compute config not found!", nil))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "OK", compute))
}

//Function for updating a compute-config
//@Summary Update a compute-config to database
//@Description Update a compute-config
//@Tags compute-config
//@Accept json
//@Product json
//@Param id path string true "ComputeConfId"
//@Param compute_config body string true "ComputeConfig"
//@Success 200 {object} entities.ComputeConfig "compute_config data with success message"
//@Failure 500 {object} nil "compute_config null with failure message"
//@Router /api/compute-config/{id} [put]
func UpdateCompute(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Compute config id is missing!", nil))
	}

	var compute entities.ComputeConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_COMPUTE, &compute); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Compute config not found!", nil))
	}

	var updateCompute entities.ComputeConfig
	if err := c.BodyParser(&updateCompute); err != nil {
		c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	utils.EntityUpdateCheck(utils.UpdateChecker, &compute, &updateCompute)
	compute.UpdatedAt = time.Now()

	database.Set(utils.KEY_PREFIX_COMPUTE+id, &compute)

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "OK", compute))
}

// Function for delete a compute-config
// @Summary Delete a compute-config from database
// @Description Delete a compute-config
// @Tags compute-config
// @Accept json
// @Product json
// @Param id path string true "ComputeConfId"
// @Success 200 {object} entities.ComputeConfig "compute-config data with success message"
// @Failure 404 {object} nil "compute-config data with null and error message"
// @Router /api/compute-config/{id} [delete]
func DeleteCompute(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Compute config id is missing!", nil))
	}

	var service entities.ComputeConfig
	if err := database.FindEntity(id, utils.KEY_PREFIX_COMPUTE, &service); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Compute config not found!", nil))
	}

	if err := database.Del(utils.KEY_PREFIX_COMPUTE + id); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Compute config has been deleted!", nil))
}
