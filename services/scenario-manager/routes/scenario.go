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
	"fmt"
	"net/http"
	"strings"
	"time"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	"github.com/futurewei-cloud/merak/services/scenario-manager/database"
	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
	"github.com/futurewei-cloud/merak/services/scenario-manager/handler"
	"github.com/futurewei-cloud/merak/services/scenario-manager/logger"
	"github.com/futurewei-cloud/merak/services/scenario-manager/utils"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/encoding/protojson"
)

//Function for doing some actions for a scenario
//@Summary Do something on a scenario
//@Description Take an action on a scenario
//@Tags scenario
//@Accept json
//@Product json
//@Param scenario body entities.ScenarioAction true "ScenarioAction"
//@Success 200 {object} entities.ScenarioAction "scenario action data with success message"
//@Failure 500 {object} nil "scenario action null with failure message"
//@Router /api/scenarios/actions [post]
func ScenarioActoins(c *fiber.Ctx) error {
	var scenarioAction entities.ScenarioAction

	if err := c.BodyParser(&scenarioAction); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	var scenario entities.Scenario
	if err := database.FindEntity(scenarioAction.ScenarioId, utils.KEY_PREFIX_SCENARIO, &scenario); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Scenario not found!", nil))
	}

	if scenario.Status != entities.STATUS_DONE && scenario.Status != entities.STATUS_NONE && scenario.Status != entities.STATUS_FAILED {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Scenario is not availalbe now!", nil))
	}

	if err := checkRelatedEntities(&scenario); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	scenario.Status = entities.STATUS_DEPLOYING
	database.Set(utils.KEY_PREFIX_SCENARIO+scenario.Id, &scenario)

	var returnBody interface{}
	var scenarioStatus entities.ServiceStatus
	var returnMessage string

	if strings.ToLower(scenarioAction.Service.ServiceName) == "topology" {
		returnTopo, err := handler.TopologyHandler(&scenario, scenarioAction.Service.Action)
		if err != nil || returnTopo.ReturnCode == pb.ReturnCode_FAILED {
			scenarioStatus = entities.STATUS_FAILED
			logger.Log.Errorf("'%s' topology failed: %s", scenarioAction.Service.Action, err.Error())
		} else {
			if scenarioAction.Service.Action == entities.EVENT_DELETE {
				scenarioStatus = entities.STATUS_NONE
			} else {
				scenarioStatus = entities.STATUS_DONE
			}
			logger.Log.Infof("'%s' topology done.", scenarioAction.Service.Action)
			logger.Log.Infof("returnTopo for action %s : %s", scenarioAction.Service.Action, returnTopo)

			var ready = 0
			var deplolying = 0
			var deleting = 0
			var errors = 0
			var done = 0
			var others = 0
			for _, cn := range returnTopo.GetComputeNodes() {
				switch cn.Status {
				case pb.Status_DONE:
					done++
				case pb.Status_READY:
					ready++
				case pb.Status_DEPLOYING:
					deplolying++
				case pb.Status_DELETING:
					deleting++
				case pb.Status_ERROR:
					errors++
				default:
					others++
				}
			}
			returnMessage = fmt.Sprintf("%s on %s got - DONE: %d, READY: %d, DEPLOYING: %d, DELETING: %d, ERROR: %d, Others: %d", scenarioAction.Service.Action, "Topology", done, ready, deplolying, deleting, errors, others)

			ret, err := protojson.Marshal(returnTopo)
			if err != nil {
				logger.Log.Errorf("returnBody Error: %s", err.Error())
			} else {
				if err1 := json.Unmarshal(ret, &returnBody); err1 != nil {
					logger.Log.Errorf("Unmarshal error: %s", err1.Error())
				}
			}
		}
	} else if strings.ToLower(scenarioAction.Service.ServiceName) == "network" {
		returnNetwork, err := handler.NetworkHandler(&scenario, scenarioAction.Service.Action)
		if err != nil || returnNetwork.ReturnCode == pb.ReturnCode_FAILED {
			scenarioStatus = entities.STATUS_FAILED
			logger.Log.Errorf("'%s' network failed: %s", scenarioAction.Service.Action, err.Error())
		} else {
			scenarioStatus = entities.STATUS_DONE
			logger.Log.Infof("'%s' network done.", scenarioAction.Service.Action)
			logger.Log.Infof("returnNetwork for action %s : %s", scenarioAction.Service.Action, returnNetwork)

			returnMessage = fmt.Sprintf("%s on %s done", scenarioAction.Service.Action, "Network")

			ret, err := protojson.Marshal(returnNetwork)
			if err != nil {
				logger.Log.Errorf("returnBody Error: %s", err.Error())
			} else {
				if err1 := json.Unmarshal(ret, &returnBody); err1 != nil {
					logger.Log.Errorf("Unmarshal error: %s", err1.Error())
				}
			}
		}
	} else if strings.ToLower(scenarioAction.Service.ServiceName) == "compute" {
		returnCompute, err := handler.ComputeHanlder(&scenario, scenarioAction.Service.Action)
		if err != nil || returnCompute.ReturnCode == pb.ReturnCode_FAILED {
			scenarioStatus = entities.STATUS_FAILED
			logger.Log.Errorf("'%s' compute failed: %s", scenarioAction.Service.Action, err.Error())
		} else {
			scenarioStatus = entities.STATUS_DONE
			logger.Log.Infof("'%s' compute done.", scenarioAction.Service.Action)
			logger.Log.Infof("returnCompute for action %s : %s", scenarioAction.Service.Action, returnCompute)

			var ready = 0
			var deplolying = 0
			var deleting = 0
			var errors = 0
			var done = 0
			var others = 0
			for _, vm := range returnCompute.GetVms() {
				switch vm.Status {
				case pb.Status_DONE:
					done++
				case pb.Status_READY:
					ready++
				case pb.Status_DEPLOYING:
					deplolying++
				case pb.Status_DELETING:
					deleting++
				case pb.Status_ERROR:
					errors++
				default:
					others++
				}
			}
			returnMessage = fmt.Sprintf("%s on %s got - DONE: %d, READY: %d, DEPLOYING: %d, DELETING: %d, ERROR: %d, Others: %d", scenarioAction.Service.Action, "Compute", done, ready, deplolying, deleting, errors, others)

			ret, err := protojson.Marshal(returnCompute)
			if err != nil {
				logger.Log.Errorf("returnBody Error: %s", err.Error())
			} else {
				if err1 := json.Unmarshal(ret, &returnBody); err1 != nil {
					logger.Log.Errorf("Unmarshal error: %s", err1.Error())
				}
			}
		}
	} else {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Scenario Action Failed.", scenarioAction))
	}

	scenario.Status = scenarioStatus
	scenario.UpdatedAt = time.Now()
	database.Set(utils.KEY_PREFIX_SCENARIO+scenario.Id, &scenario)

	if scenario.Status == entities.STATUS_FAILED {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", "Scenario Action Failed.", scenarioAction))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Action successfully - "+returnMessage, returnBody))
}

//Function for creating a scenario
//@Summary Insert a scenario to database
//@Description Create a scenario
//@Tags scenario
//@Accept json
//@Product json
//@Param scenario body entities.Scenario true "Scenario"
//@Success 200 {object} entities.Scenario "scenario data with success message"
//@Failure 500 {object} nil "scenario null with failure message"
//@Router /api/scenarios [post]
func CreateScenario(c *fiber.Ctx) error {
	var scenario entities.Scenario

	if err := c.BodyParser(&scenario); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	var id = utils.GenUUID()
	scenario.Id = id

	if err := checkRelatedEntities(&scenario); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	scenario.Status = entities.STATUS_NONE
	scenario.CreatedAt = time.Now()
	scenario.UpdatedAt = time.Now()

	database.Set(utils.KEY_PREFIX_SCENARIO+id, &scenario)

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Scenario Has been created successfully.", scenario))
}

//Function for retriving all scenarios
//@Summary Get all scenario from database
//@Description Get all scenario
//@Tags scenario
//@Accept json
//@Product json
//@Success 200 {object} []entities.Scenario "array of scenario with success message"
//@Failure 404 {object} nil "null scenario data with error message"
//@Router /api/senarios [get]
func GetScenarios(c *fiber.Ctx) error {
	var values map[string]string

	values, err := database.GetAllValuesWithKeyPrefix(utils.KEY_PREFIX_SCENARIO)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	if len(values) < 1 {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", errors.New("scenario not present").Error(), nil))
	}

	var responseScenarios []entities.Scenario

	for _, value := range values {
		var scenario entities.Scenario

		_ = json.Unmarshal([]byte(value), &scenario)
		responseScenarios = append(responseScenarios, scenario)
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"status": "OK", "message": "OK", "data": responseScenarios})
}

//Function for retriving a scenario
//@Summary Get a scenario from database
//@Description Get a scenario
//@Tags scenario
//@Accept json
//@Product json
//@Param id path string true "ScenarioId"
//@Success 200 {object} entities.Scenario "scenario data with success message"
//@Failure 404 {object} nil "scenario data with null and error message"
//@Router /api/senarios/{id} [get]
func GetScenario(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Scenario id is missing!", nil))
	}

	var scenario entities.Scenario
	if err := database.FindEntity(id, utils.KEY_PREFIX_SCENARIO, &scenario); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Scenario not found!", nil))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "OK", scenario))
}

//Function for updating a scenario
//@Summary Update a scenario to database
//@Description Update a scenario
//@Tags scenario
//@Accept json
//@Product json
//@Param id path string true "ScenarioId"
//@Param scenario body string true "Scenario"
//@Success 200 {object} entities.Scenario "scenario data with success message"
//@Failure 500 {object} nil "scenario null with failure message"
//@Router /api/senarios/{id} [put]
func UpdateScenario(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Scenario id is missing!", nil))
	}

	var scenario entities.Scenario
	if err := database.FindEntity(id, utils.KEY_PREFIX_SCENARIO, &scenario); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Scenario not found!", nil))
	}

	var updateScenario entities.Scenario
	if err := c.BodyParser(&updateScenario); err != nil {
		c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	utils.EntityUpdateCheck(utils.UpdateChecker, &scenario, &updateScenario)
	scenario.UpdatedAt = time.Now()

	if err := checkRelatedEntities(&scenario); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	database.Set(utils.KEY_PREFIX_SCENARIO+id, &scenario)

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "OK", scenario))
}

// Function for delete a scenario
// @Summary Delete a scenario from database
// @Description Delete a scenario
// @Tags scenario
// @Accept json
// @Product json
// @Param id path string true "ScenarioId"
// @Success 200 {object} entities.Scenario "scenario data with success message"
// @Failure 404 {object} nil "scenario data with null and error message"
// @Router /api/senarios/{id} [delete]
func DeleteScenario(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ReturnResponseMessage("FAILED", "Scenario id is missing!", entities.Scenario{}))
	}

	var scenario entities.Scenario
	if err := database.FindEntity(id, utils.KEY_PREFIX_SCENARIO, &scenario); err != nil {
		return c.Status(http.StatusNotFound).JSON(utils.ReturnResponseMessage("FAILED", "Scenario not found!", nil))
	}

	if err := database.Del(utils.KEY_PREFIX_SCENARIO + id); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ReturnResponseMessage("FAILED", err.Error(), nil))
	}

	return c.Status(http.StatusOK).JSON(utils.ReturnResponseMessage("OK", "Scenario has been deleted!", nil))
}

func checkRelatedEntities(scenario *entities.Scenario) error {
	var topology entities.TopologyConfig
	if err := database.FindEntity(scenario.TopologyId, utils.KEY_PREFIX_TOPOLOGY, &topology); err != nil {
		return errors.New("topology not found")
	}

	var service entities.ServiceConfig
	if err := database.FindEntity(scenario.ServiceConfId, utils.KEY_PREFIX_SERVICE, &service); err != nil {
		return errors.New("service config not found")
	}

	var network entities.NetworkConfig
	if err := database.FindEntity(scenario.NetworkConfId, utils.KEY_PREFIX_NETWORK, &network); err != nil {
		return errors.New("network config not found")
	}

	var compute entities.ComputeConfig
	if err := database.FindEntity(scenario.ComputeConfId, utils.KEY_PREFIX_COMPUTE, &compute); err != nil {
		return errors.New("compute config not found")
	}

	var test entities.TestConfig
	if err := database.FindEntity(scenario.TestConfId, utils.KEY_PREFIX_TEST, &test); err != nil {
		return errors.New("test config not found")
	}

	return nil
}
