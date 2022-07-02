package utils

import (
	"github.com/google/uuid"
	"strings"
)

func GenUUID() string {
	//uuidWithHyphen, _ := uuid.NewRandom()
	return strings.Replace(uuid.New().String(), "-", "", -1)
}
func ReturnError() string {

	return ""
}
