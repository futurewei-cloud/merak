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

package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/futurewei-cloud/merak/services/scenario-manager/entities"
	"github.com/futurewei-cloud/merak/services/scenario-manager/logger"
	"github.com/go-redis/redis/v8"
)

var (
	ErrNil = errors.New("no matching record found in redis database")
	Ctx    = context.Background()
	Rdb    *redis.Client
)

func ConnectDatabase(cfg *entities.AppConfig) error {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.DBHost + ":" + cfg.DBPort,
		Username: cfg.DBUser,
		Password: cfg.DBPass,
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
		logger.Log.Errorf("database SET %s VALUE %s failed %s", key, jsonVal, err.Error())
		return err
	}
	return nil
}

func Get(key string) (string, error) {
	val, err := Rdb.Get(Ctx, key).Result()
	if err != nil {
		logger.Log.Errorf("database GET %s failed %s", key, err)
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

	iter := Rdb.Scan(Ctx, 0, prefix, 0).Iterator()
	for iter.Next(Ctx) {
		allkeys = append(allkeys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		logger.Log.Errorf("database scan keys %s failed %s", prefix, err)
		return nil, fmt.Errorf("scan db error '%s' when retriving key '%s' keys", err, prefix)
	}

	return allkeys, nil
}

func getKeyAndValueMap(keys []string, prefix string) (map[string]string, error) {
	values := make(map[string]string)
	for _, key := range keys {
		value, err := Rdb.Get(Ctx, key).Result()
		if err != nil {
			logger.Log.Errorf("database scan keys %s failed %s", prefix, err)
			return nil, fmt.Errorf("get value error '%s' when retriving key '%s' keys", err, prefix)
		}

		// Strip off the prefix from the key so that we save the key to the value
		strippedKey := strings.Split(key, prefix)
		values[strippedKey[1]] = value
	}
	return values, nil
}
