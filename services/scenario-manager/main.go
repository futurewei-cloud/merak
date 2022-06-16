package main

import (
	"fmt"
	"log"

	"github.com/futurewei-cloud/merak/services/scenario-manager/database"
	_ "github.com/futurewei-cloud/merak/services/scenario-manager/docs"
	"github.com/futurewei-cloud/merak/services/scenario-manager/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

// @title Merak Rest API
// @version 2.0
// @description Restful APIs for Merak Cloud Emulator
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /
// @schemes http
func main() {
	app := Setup()

	// Start Server
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}

func welcome(c *fiber.Ctx) error {
	return c.SendString("Welcome to Merak - Cloud Emulator")
}

func Setup() *fiber.App {
	// Connect to storage
	if err := database.ConnectDatabase(); err != nil {
		fmt.Printf("Cannot connect to Redis db!, error: '%s'\n", err)
	}

	// Fiber instance
	app := fiber.New()

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New())

	// Routes
	SetupRoutes(app)

	return app
}

// function making for setting internal routes
func SetupRoutes(app *fiber.App) {
	apiURL := "/api"

	// Welcome endpoint - not required Autorization
	app.Get(apiURL, welcome)

	// Public resource in static route
	//app.Static("/", "./public")
	// Register the index route with a simple
	// "OK" response. It should return status
	// code 200
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	//app.Get("/", HealthCheck)
	app.Get("/swagger/*", swagger.HandlerDefault) // default

	// User Endpoints
	//user := app.Group(apiURL + "/users")
	//user.Post("/login", routes.Auth)
	//user.Post("/", routes.CreateUser)

	// AuthorizationRequired Action
	//app.Use(routes.AuthorizationRequired())

	// need AutorizationRequired
	//user.Get("/", routes.GetUsers)
	//user.Get("/:id", routes.GetUser)
	//user.Put("/:id", routes.UpdateUser)
	//user.Delete("/:id", routes.DeleteUser)

	// Scenario
	scenario := app.Group(apiURL + "/scenarios")
	scenario.Post("/", routes.CreateScenario)
	scenario.Get("/", routes.GetScenarios)
	scenario.Get("/:id", routes.GetScenario)
	scenario.Put("/:id", routes.UpdateScenario)
	scenario.Delete("/:id", routes.DeleteScenario)
	scenario.Post("/action", routes.ScenarioActoins)

	// Topology
	topology := app.Group(apiURL + "/topologies")
	topology.Post("/", routes.CreateTopology)
	topology.Get("/", routes.GetTopologies)
	topology.Get("/:id", routes.GetTopology)
	topology.Put("/:id", routes.UpdateTopology)
	topology.Delete("/:id", routes.DeleteTopology)

	// Service-config
	service := app.Group(apiURL + "/service-config")
	service.Post("/", routes.CreateService)
	service.Get("/", routes.GetServices)
	service.Get("/:id", routes.GetService)
	service.Put("/:id", routes.UpdateService)
	service.Delete("/:id", routes.DeleteService)

	// Network-config
	network := app.Group(apiURL + "/network-config")
	network.Post("/", routes.CreateNetwork)
	network.Get("/", routes.GetNetworks)
	network.Get("/:id", routes.GetNetwork)
	network.Put("/:id", routes.UpdateNetwork)
	network.Delete("/:id", routes.DeleteNetwork)

	// Compute-config
	compute := app.Group(apiURL + "/compute-config")
	compute.Post("/", routes.CreateCompute)
	compute.Get("/", routes.GetComputes)
	compute.Get("/:id", routes.GetCompute)
	compute.Put("/:id", routes.UpdateCompute)
	compute.Delete("/:id", routes.DeleteCompute)

	// Test-config
	test := app.Group(apiURL + "/test-config")
	test.Post("/", routes.CreateTestConfig)
	test.Get("/", routes.GetTestConfigs)
	test.Get("/:id", routes.GetTestConfig)
	test.Put("/:id", routes.UpdateTestConfig)
	test.Delete("/:id", routes.DeleteTestConfig)

	//end AuthorizatoinRequired
}

// HealthCheck godoc
// @Summary Show the status of server.
// @Description get the status of server.
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func HealthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	if err := c.JSON(res); err != nil {
		return err
	}

	return nil
}
