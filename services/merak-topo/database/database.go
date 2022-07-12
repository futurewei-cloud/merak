package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
	corev1 "k8s.io/api/core/v1"
)

var (
	ErrNil = errors.New("no matching record found in redis database")
	Ctx    = context.Background()
	Rdb    *redis.Client
)

func ConnectDatabase() error {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       2,  // use default DB
	})

	if err := client.Ping(Ctx).Err(); err != nil {
		return err
	}

	Rdb = client
	return nil
}

func SetValue(key string, val interface{}) error {
	j, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("fail to save value in DB %s", err)
	}
	err = Rdb.Set(Ctx, key, j, 0).Err()
	if err != nil {
		return fmt.Errorf("fail to save value in DB %s", err)
	}

	return nil
}

func Get(key string) (string, error) {
	val, err := Rdb.Get(Ctx, key).Result()
	if err != nil {
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
func FindEntity(id string, prefix string, entity interface{}) (interface{}, error) {
	if (id + prefix) == "" {
		return "invalid input", nil
	}
	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return "fail to get value for key in DB", err
	}
	err = json.Unmarshal([]byte(value), &entity)
	if err != nil {
		return "fail to unmarshal value in DB", err
	}
	return entity, nil
}

func FindPodEntity(id string, prefix string) (*corev1.Pod, error) {
	var entity *corev1.Pod

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(value), &entity)
	if err != nil {
		panic(err)
	}
	return entity, nil
}

func FindTopoEntity(id string, prefix string) (TopologyData, error) {
	var entity TopologyData

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return entity, fmt.Errorf("fail to get value for key in DB %s", err)
		// panic(err)
	}
	err = json.Unmarshal([]byte(value), &entity)
	if err != nil {
		// return fmt.Errorf("fail to get value for key in DB %s", err)
		panic(err)
	}
	return entity, nil
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

func DeleteAllValuesWithKeyPrefix(prefix string) error {
	keys, err := getKeys(fmt.Sprintf("%s*", prefix))
	if err != nil {
		return fmt.Errorf("fail to get keys with the prefix %s", err)
	}

	for _, key := range keys {

		err := Del(key)
		if err != nil {
			return fmt.Errorf("fail to remove key %v in DB %s", key, err)
		}

	}

	return nil
}

func getKeys(prefix string) ([]string, error) {
	var allkeys []string

	iter := Rdb.Scan(Ctx, 0, prefix, 0).Iterator()
	for iter.Next(Ctx) {
		allkeys = append(allkeys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("scan db error '%s' when retriving key '%s' keys", err, prefix)
	}

	return allkeys, nil
}
func getKeyAndValueMap(keys []string, prefix string) (map[string]string, error) {
	values := make(map[string]string)
	for _, key := range keys {
		value, err := Rdb.Get(Ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("Get value error '%s' when retriving key '%s' keys", err, prefix)
		}

		// Strip off the prefix from the key so that we save the key to the value
		strippedKey := strings.Split(key, prefix)
		values[strippedKey[1]] = value
	}
	return values, nil
}

// pod:uid     pod info
// topo:topo_id     topoinfo
