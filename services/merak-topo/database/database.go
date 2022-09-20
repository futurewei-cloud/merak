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
	"log"

	"strings"

	"github.com/golang/protobuf/proto"

	pb_common "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/topology"
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

		Addr:     "topology-redis-master.merak.svc.cluster.local:55001",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err := client.Ping(Ctx).Err(); err != nil {
		return fmt.Errorf("ConnectDB: connect DB error %s", err.Error())
	}

	Rdb = client
	return nil
}

func SetValue(key string, val interface{}) error {
	j, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("SetValue: save value json marshal error %s", err.Error())
	}
	err2 := Rdb.Set(Ctx, key, j, 0).Err()
	if err2 != nil {
		return fmt.Errorf("SetValue: save value in DB error %s", err2.Error())
	}

	return nil
}

func SetPbReturnValue(key string, val *pb.ReturnTopologyMessage) error {
	j, err := proto.Marshal(val)
	if err != nil {
		return fmt.Errorf("SetPbReturnValue: proto marshal error %s", err.Error())
	}
	err2 := Rdb.Set(Ctx, key, j, 0).Err()
	if err2 != nil {
		return fmt.Errorf("SetPbReturnValue: save value in DB %s", err2.Error())
	}

	return nil
}

func GetPbReturnValue(id string, prefix string, entity *pb.ReturnTopologyMessage) error {

	if (id + prefix) == "" {
		return fmt.Errorf("GetPbReturnValue: get key is empty")
	}
	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return fmt.Errorf("GetPbReturnValue: get value for key in DB %s", err.Error())
	}
	err2 := proto.Unmarshal([]byte(value), entity)
	if err2 != nil {
		return fmt.Errorf("GetPbReturnValue: unmarshal error %s", err2.Error())
	}
	return nil
}

func Get(key string) (string, error) {
	val, err := Rdb.Get(Ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("Get: get from DB error %s", err.Error())
	}
	return val, nil
}

func Del(key string) error {
	if err := Rdb.Del(Ctx, key).Err(); err != nil {
		return fmt.Errorf("Del: delete error %s", err.Error())
	}
	return nil
}

// Function for finding an entity from database
func FindEntity(id string, prefix string, entity interface{}) (interface{}, error) {
	if (id + prefix) == "" {
		log.Printf("FindEntity:invalid input of entity")
	}
	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return "", fmt.Errorf("FindEntity: get value for key in DB %s", err.Error())
	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		return "", fmt.Errorf("FindEntity: unmarshal key error %s", err2.Error())
	}
	return entity, nil
}

func FindIPEntity(id string, prefix string) ([]string, error) {
	var entity []string

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return nil, fmt.Errorf("FindIPEntity: get value from DB error %s", err.Error())
	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		return nil, fmt.Errorf("FindIPEntity: unmarshal error %s", err2.Error())
	}
	return entity, nil
}

func FindPodEntity(id string, prefix string) (*corev1.Pod, error) {
	var entity *corev1.Pod

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return nil, fmt.Errorf("FindPodEntity:get value from DB error %s", err.Error())
	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		return nil, fmt.Errorf("FindPodEntity:unmarshal error %s", err2.Error())
	}
	return entity, nil
}

func FindHostNode(id string, prefix string) ([]*pb_common.InternalHostInfo, error) {
	var entity []*pb_common.InternalHostInfo

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return nil, fmt.Errorf("FindHostNode:get value from DB error %s", err.Error())
	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		return nil, fmt.Errorf("FindHostNode:unmarshal error %s", err2.Error())
	}
	return entity, nil
}

func FindComputenode(id string, prefix string) ([]*pb_common.InternalComputeInfo, error) {
	var entity []*pb_common.InternalComputeInfo

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return nil, fmt.Errorf("FindComputenode:get value from DB error %s", err.Error())
	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		return nil, fmt.Errorf("FindComputenode:unmarshal error %s", err2.Error())
	}
	return entity, nil
}

func FindTopoEntity(id string, prefix string) (TopologyData, error) {
	var entity TopologyData

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return entity, fmt.Errorf("FindTopoEntity:get value from DB error %s", err.Error())

	}
	err2 := json.Unmarshal([]byte(value), &entity)
	if err2 != nil {
		return entity, fmt.Errorf("FindTopoEntity: unmarshal error %s", err2.Error())
	}
	return entity, nil
}

func GetAllValuesWithKeyPrefix(prefix string) (map[string]string, error) {
	keys, err := getKeys(fmt.Sprintf("%s*", prefix))
	if err != nil {
		return nil, fmt.Errorf("GetAllValuesWithKeyPrefix:get keys error %s", err.Error())
	}

	values, err2 := getKeyAndValueMap(keys, prefix)
	if err2 != nil {
		return nil, fmt.Errorf("GetAllValuesWithKeyPrefix:get key and value map error %s", err2.Error())
	}
	return values, nil
}

func DeleteAllValuesWithKeyPrefix(prefix string) error {
	keys, err := getKeys(fmt.Sprintf("%s*", prefix))
	if err != nil {
		return fmt.Errorf("DeleteAllValuesWithKeyPrefix: get keys error %s", err.Error())
	}

	for _, key := range keys {

		err2 := Del(key)
		if err2 != nil {
			return fmt.Errorf("DeleteAllValuesWithKeyPrefix: Del key %v error %s", key, err2.Error())
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
		return nil, fmt.Errorf("getKeys:scan db error '%s' when retriving key '%s' keys", err.Error(), prefix)
	}

	return allkeys, nil
}
func getKeyAndValueMap(keys []string, prefix string) (map[string]string, error) {
	values := make(map[string]string)
	for _, key := range keys {
		value, err := Rdb.Get(Ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("getKeyAndValueMap:Get value error '%s' when retriving key '%s' keys", err.Error(), prefix)
		}

		strippedKey := strings.Split(key, prefix)
		values[strippedKey[1]] = value
	}
	return values, nil
}
