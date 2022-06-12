package database

import (
	"log"
	"os"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
)

var (
	client *redis.Client
)
var (
	key = "key"
	val = "val"
)

func TestDBMain(m *testing.M) {
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	client = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	code := m.Run()
	os.Exit(code)
}
