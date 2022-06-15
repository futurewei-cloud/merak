package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()
var RDB *redis.Client

func CreateDBClient() *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return client
}

func PingClient(client *redis.Client) error {
	pong, err := client.Ping(Ctx).Result()
	if err != nil {
		return err
	}
	fmt.Println(pong, err)
	return nil
}

func SetValue(client *redis.Client, key string, val interface{}) {
	j, err := json.Marshal(val)
	if err != nil {
		fmt.Println("error:", err)
	}
	err2 := client.Set(Ctx, key, j, 0).Err()
	if err2 != nil {
		// panic(err2)
		fmt.Println(err2.Error())
	} else {
		fmt.Printf("Save data in Redis")
	}
}

func GetValue(client *redis.Client, key string) string {
	val, err := client.Get(Ctx, key).Result()
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("Get data in Redis")
	}
	return val
}
