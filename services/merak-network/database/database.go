package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/go-redis/redis/v8"
)

var (
	ErrNil = errors.New("no matching record found in redis database")
	Ctx    = context.Background()
	Rdb    *redis.Client
)

func ConnectDatabase() error {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:55000",
		Username: "default",
		Password: "redispw",
		DB:       0,
	})

	if err := client.Ping(Ctx).Err(); err != nil {
		return err
	}

	Rdb = client
	return nil
}

func Set(key string, val interface{}) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = Rdb.Set(Ctx, key, jsonVal, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func Get(key string) (string, error) {
	log.Printf("DB GET %s", key)
	val, err := Rdb.Get(Ctx, key).Result()
	if err != nil {
		log.Println("DB Get Issue")
		return "", err
	}
	return val, nil
}

func Del(key string) error {
	if err := Rdb.Del(Ctx, key).Err(); err != nil {
		return err
	}
	return nil
}

// Function for finding an entity from database
func FindEntity(id string, prefix string, entity interface{}) error {
	if id == "" {
		return errors.New("invalid for id parameter")
	}
	value, err := Rdb.Get(Ctx, prefix+id).Result()
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(value), &entity)
	return err
}

func GetAllValuesWithKeyPrefix(prefix string) (map[string]string, error) {
	keys, err := getKeys(fmt.Sprintf("%s*", prefix))
	if err != nil {
		return nil, err
	}

	values, err := getKeyAndValueMap(keys, prefix)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func getKeys(prefix string) ([]string, error) {
	var allkeys []string
	var cursor uint64
	count := int64(10)

	for {
		var keys []string
		var err error
		keys, cursor, err := Rdb.Scan(Ctx, cursor, prefix, count).Result()
		if err != nil {
			return nil, fmt.Errorf("scan db error '%s' when retriving key '%s' keys", err, prefix)
		}

		allkeys = append(allkeys, keys...)
		if cursor == 0 {
			break
		}
	}
	return allkeys, nil
}

func getKeyAndValueMap(keys []string, prefix string) (map[string]string, error) {
	values := make(map[string]string)
	for _, key := range keys {
		value, err := Rdb.Get(Ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("get value error '%s' when retriving key '%s' keys", err, prefix)
		}

		// Strip off the prefix from the key so that we save the key to the value
		strippedKey := strings.Split(key, prefix)
		values[strippedKey[1]] = value
	}
	return values, nil
}
