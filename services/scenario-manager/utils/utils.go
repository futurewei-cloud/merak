package utils

import (
	"reflect"
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

func UpdateChecker(src interface{}, upt interface{}) interface{} {
	switch src := src.(type) {
	case string:
		if upt != "" {
			return upt
		}
		return src
	}
	return src
}

func EntityUpdateCheck(check func(interface{}, interface{}) interface{}, origin interface{}, update interface{}) {
	src := reflect.ValueOf(origin).Elem()
	upt := reflect.ValueOf(update).Elem()
	for i := 0; i < src.NumField(); i++ {
		r := check(src.Field(i).Interface(), upt.Field(i).Interface())
		src.Field(i).Set(reflect.ValueOf(r))
	}
}
