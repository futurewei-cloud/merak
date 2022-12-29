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

	// "strings"

	"github.com/futurewei-cloud/merak/services/merak-topo/utils"

	"github.com/go-redis/redis/v8"
)

var (
	Err_query    = errors.New("no matching record found in redis database")
	Err_setvalue = errors.New("fails to save data in redis database")
	Err_delete   = errors.New("fails to delete data in redis database")
	Ctx          = context.Background()
	Rdb          *redis.Client
)

func ConnectDatabase() error {
	client := redis.NewClient(&redis.Options{

		Addr:     "topology-redis-master.merak.svc.cluster.local:55001",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err := client.Ping(Ctx).Err(); err != nil {
		utils.Logger.Error("can't connect to DB", "connect DB error", err.Error())
		return err
	}

	Rdb = client
	return nil
}

func SetValue(key string, val interface{}) error {

	err_flag := 0

	j, err := json.Marshal(val)
	if err != nil {
		utils.Logger.Error("can't marshal ", "json marshal ", err.Error())
		err_flag = 1
	}
	err2 := Rdb.Set(Ctx, key, j, 0).Err()
	if err2 != nil {
		utils.Logger.Error("can't save key in DB", "error ", err2.Error(), "key", key)
		err_flag = 1
	}

	if err_flag == 1 {
		return Err_setvalue
	} else {
		return nil
	}

}

func Get(key string) (string, error) {
	val, err := Rdb.Get(Ctx, key).Result()
	if err != nil {
		utils.Logger.Error("can't get value from DB", "error ", err.Error(), "key", key)
		return val, err
	}
	return val, nil
}

func Del(key string) error {
	if err := Rdb.Del(Ctx, key).Err(); err != nil {
		utils.Logger.Error("can't delete key in DB", "error", err.Error(), "key", key)
		return err
	}
	return nil
}

// Function for finding an entity from database
func FindEntity(id string, prefix string, entity interface{}) error {

	errs_flag := 0

	if (id + prefix) == "" {
		utils.Logger.Error("can't find entity", "entity key", "empty key")
		errs_flag = 1
	}

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		utils.Logger.Error("can't find entity", id+prefix, err.Error())
		errs_flag = 1
	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		utils.Logger.Error("can't find entity", "unmarshal key ", err2.Error())
		errs_flag = 1
	}

	if errs_flag == 1 {
		return Err_query
	} else {
		return nil
	}

}

func FindHostEntity(id string, prefix string) (HostNode, error) {
	var entity HostNode

	errs_flag := 0
	var err_return error

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		utils.Logger.Error("can't find host entity", id+prefix, err.Error())
		errs_flag = 1
	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		utils.Logger.Error("can't find host entity", "unmarshal", err2.Error())
		errs_flag = 1
	}

	if errs_flag == 1 {
		err_return = Err_query
	} else {
		err_return = nil
	}
	return entity, err_return
}

func FindComputeEntity(id string, prefix string) (ComputeNode, error) {
	var entity ComputeNode
	errs_flag := 0
	var err_return error

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		utils.Logger.Error("can't find compute entity", id+prefix, err.Error())
		errs_flag = 1
	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		utils.Logger.Error("can't find compute entity", value, err2.Error())
		errs_flag = 1
	}
	if errs_flag == 1 {
		err_return = Err_query
	} else {
		err_return = nil
	}
	return entity, err_return
}

/*Comment: unused function*/
// func GetAllValuesWithKeyPrefix(prefix string) (map[string]string, error) {
// 	keys, err := getKeys(fmt.Sprintf("%s*", prefix))
// 	if err != nil {
// 		utils.Logger.Error("GetAllValuesWithKeyPrefix:get keys error", err.Error())
// 	}

// 	values, err2 := getKeyAndValueMap(keys, prefix)
// 	if err2 != nil {
// 		utils.Logger.Error("GetAllValuesWithKeyPrefix:get key and value map error ", err2.Error())
// 		return values, err2
// 	}
// 	return values, nil

// }

func DeleteAllValuesWithKeyPrefix(prefix string) error {
	err_flag := 0
	keys, err := getKeys(fmt.Sprintf("%s*", prefix))
	if err != nil {
		utils.Logger.Error("can't find prefix value in DB", prefix, err.Error())
		err_flag = 1
	}

	for _, key := range keys {
		err2 := Del(key)
		if err2 != nil {
			utils.Logger.Error("can't delete key in DB", key, err2.Error())
			err_flag = 1
		}
	}

	if err_flag == 1 {
		return Err_delete
	} else {
		return nil
	}
}

func getKeys(prefix string) ([]string, error) {
	var allkeys []string

	var err_return error

	iter := Rdb.Scan(Ctx, 0, prefix, 0).Iterator()
	for iter.Next(Ctx) {
		allkeys = append(allkeys, iter.Val())
	}

	err := iter.Err()
	if err != nil {
		utils.Logger.Error("can't iterate in DB", prefix, err.Error())
		err_return = err
	} else {
		err_return = nil
	}

	return allkeys, err_return
}

/*Comment: unused function*/

// func getKeyAndValueMap(keys []string, prefix string) (map[string]string, error) {
// 	values := make(map[string]string)
// 	for _, key := range keys {
// 		value, err := Rdb.Get(Ctx, key).Result()
// 		if err != nil {
// 			utils.Logger.Error("getKeyAndValueMap:Get value error when retriving keys", err.Error(), prefix)
// 			return nil, err
// 		}

// 		strippedKey := strings.Split(key, prefix)
// 		values[strippedKey[1]] = value
// 	}
// 	return values, nil
// }

func FindTopoEntity(id string, prefix string) (TopologyData, error) {
	var entity TopologyData

	err_flag := 0
	var err_return error

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		utils.Logger.Error("can't find topology entity", id+prefix, err.Error())
		err_flag = 1
	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		utils.Logger.Error("can't find topology entity", "unmarshal", err2.Error())
		err_flag = 1
	}
	if err_flag == 1 {
		err_return = Err_query
	} else {
		err_return = nil
	}
	return entity, err_return
}
