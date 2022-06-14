package utils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GenUUID() string {
	uuidWithHyphen, _ := uuid.NewRandom()
	return strings.Replace(uuidWithHyphen.String(), "-", "", -1)
}

func ReturnResponseMessage(status string, message string, data interface{}) map[string]interface{} {
	// if data.Id == "" {
	// 	return fiber.Map{"status": status, "message": message, "data": nil}
	// }
	return fiber.Map{"status": status, "message": message, "data": data}
}
