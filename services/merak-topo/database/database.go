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
	err_query = errors.New("Invalid Input")
	Ctx       = context.Background()
	Rdb       *redis.Client
)

func ConnectDatabase() error {
	client := redis.NewClient(&redis.Options{

		Addr:     "topology-redis-master.merak.svc.cluster.local:55001",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err := client.Ping(Ctx).Err(); err != nil {
		utils.Logger.Error("can't connect to DB, please retry", "connect DB error", err.Error())
		return err
	}

	Rdb = client
	return nil
}

func SetValue(key string, val interface{}) error {
	j, err := json.Marshal(val)
	if err != nil {
		utils.Logger.Error("can't marshal, please retry ", "json marshal ", err.Error())
		return err
	}
	err2 := Rdb.Set(Ctx, key, j, 0).Err()
	if err2 != nil {
		utils.Logger.Warn("can't save key in DB, please retry", "Warning ", err2.Error(), "key", key)

	}

	return nil
}

func Get(key string) (string, error) {

	val := "XXX"

	val, err := Rdb.Get(Ctx, key).Result()
	if err != nil {
		utils.Logger.Warn("can't get key in DB, please retry", "Warning", err.Error(), "key", key)

	}

	return val, nil
}

func Del(key string) error {

	if err := Rdb.Del(Ctx, key).Err(); err != nil {
		utils.Logger.Warn("can't delete key in DB, please retry", "Warning", err.Error(), "key", key)
		return err
	}
	return nil
}

// Function for finding an entity from database
func FindEntity(id string, prefix string, entity interface{}) error {

	if (id + prefix) == "" {
		utils.Logger.Warn("can't find entity, please retry", "entity key", "empty key")
		return err_query
	}

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		utils.Logger.Warn("can't find entity, please retry", id+prefix, err.Error())

	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		utils.Logger.Warn("can't find entity, please retry", "unmarshal key ", err2.Error())

	}
	return nil
}

func FindHostEntity(id string, prefix string) (HostNode, error) {
	var entity HostNode
	entity.Ip = "X.X.X.X"
	entity.Status = STATUS_NONE
	entity.Routing_rule = []string{"Initial"}

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		utils.Logger.Warn("can't find host entity, please retry", id+prefix, err.Error())

	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		utils.Logger.Warn("can't find host entity, please retry", "unmarshal", err2.Error())

	}

	return entity, nil
}

func FindComputeEntity(id string, prefix string) (ComputeNode, error) {
	var entity ComputeNode
	entity.ContainerIp = "X.X.X.X"
	entity.DatapathIp = "X.X.X.X"
	entity.Id = "XXX"
	entity.Mac = "XXXXX"
	entity.Name = "Initial"
	entity.Status = STATUS_NONE
	entity.Veth = "Initial"

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		utils.Logger.Warn("can't find compute entity, please retry", id+prefix, err.Error())

	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		utils.Logger.Warn("can't find compute entity, please retry", value, err2.Error())

	}
	return entity, nil
}

func FindTopoEntity(id string, prefix string) (TopologyData, error) {
	var entity TopologyData
	entity.Topology_id = "XXX"

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		utils.Logger.Warn("can't find topology entity, please retry", id+prefix, err.Error())

	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		utils.Logger.Warn("can't find topology entity, please retry", "unmarshal", err2.Error())

	}

	return entity, nil
}

func getKeys(prefix string) ([]string, error) {
	allkeys := []string{"XXX"}

	iter := Rdb.Scan(Ctx, 0, prefix, 0).Iterator()
	for iter.Next(Ctx) {
		allkeys = append(allkeys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		utils.Logger.Warn("can't get keys in DB scan", prefix, err.Error())

	}

	return allkeys, nil
}

func DeleteAllValuesWithKeyPrefix(prefix string) error {

	keys, err := getKeys(fmt.Sprintf("%s*", prefix))
	if err != nil {
		utils.Logger.Warn("can't find prefix value in DB, please retry", prefix, err.Error())
		return err
	}

	for _, key := range keys {
		err2 := Del(key)
		if err2 != nil {
			utils.Logger.Warn("can't delete key in DB, please retry", key, err2.Error())
			return err2
		}
	}

	return nil

}
