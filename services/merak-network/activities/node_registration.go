package activities

import (
	"context"
	"log"
)

func NodeRegister(ctx context.Context, name string) (string, error) {
	log.Println("NodeRegister")
	return "NodeRegister", nil
}
