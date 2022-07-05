package utils

import (
	"reflect"
	"strings"

	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
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
	case string, entities.ServiceStatus:
		if upt != "" {
			return upt
		}
		return src
	case int, uint, uint32:
		if upt != 0 {
			return upt
		}
		return src
	case []string:
		if len(upt.([]string)) > 0 {
			return upt
		}
		return src
	case []entities.Service:
		if len(upt.([]entities.Service)) > 0 {
			return upt
		}
		return src
	case []entities.Image:
		if len(upt.([]entities.Image)) > 0 {
			return upt
		}
		return src
	case []entities.VNode:
		if len(upt.([]entities.VNode)) > 0 {
			return upt
		}
		return src
	case []entities.VLink:
		if len(upt.([]entities.VLink)) > 0 {
			return upt
		}
		return src
	case []entities.VPCInfo:
		if len(upt.([]entities.VPCInfo)) > 0 {
			return upt
		}
		return src
	case []entities.Router:
		if len(upt.([]entities.Router)) > 0 {
			return upt
		}
		return src
	case []entities.Gateway:
		if len(upt.([]entities.Gateway)) > 0 {
			return upt
		}
		return src
	case []entities.SecurityGroup:
		if len(upt.([]entities.SecurityGroup)) > 0 {
			return upt
		}
		return src
	case []entities.Test:
		if len(upt.([]entities.Test)) > 0 {
			return upt
		}
		return src
	default:
		return src
	}
}

func EntityUpdateCheck(check func(interface{}, interface{}) interface{}, origin interface{}, update interface{}) {
	src := reflect.ValueOf(origin).Elem()
	upt := reflect.ValueOf(update).Elem()
	for i := 0; i < src.NumField(); i++ {
		r := check(src.Field(i).Interface(), upt.Field(i).Interface())
		src.Field(i).Set(reflect.ValueOf(r))
	}
}
