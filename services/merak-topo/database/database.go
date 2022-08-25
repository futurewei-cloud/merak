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

		// Addr:     "topology-redis-master.default.svc.cluster.local:55001",
		// Addr:     "172.31.28.160:55001",
		// Addr:     "54.189.190.120:55001",
		Addr:     "10.106.191.97:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err := client.Ping(Ctx).Err(); err != nil {
		return fmt.Errorf("fail to connect DB %s", err)
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

func SetPbReturnValue(key string, val *pb.ReturnTopologyMessage) error {
	j, err := proto.Marshal(val)
	if err != nil {
		return fmt.Errorf("fail to save value in DB %s", err)
	}
	err = Rdb.Set(Ctx, key, j, 0).Err()
	if err != nil {
		return fmt.Errorf("fail to save value in DB %s", err)
	}

	return nil
}

func GetPbReturnValue(id string, prefix string, entity *pb.ReturnTopologyMessage) error {

	if (id + prefix) == "" {
		log.Println("get key is empty")
		return fmt.Errorf("get key is empty")
	}
	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		log.Printf("fail to get value for key in DB %s", err.Error())
		return fmt.Errorf("fail to get value for key in DB %s", err.Error())
	}
	err = proto.Unmarshal([]byte(value), entity)
	if err != nil {
		log.Printf("fail to unmarshal in DB %s", err.Error())
		return fmt.Errorf("fail to unmarshal in DB %s", err.Error())
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

func FindIPEntity(id string, prefix string) ([]string, error) {
	var entity []string

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return nil, fmt.Errorf("fail to find pod %s", err)
	}
	err = json.Unmarshal([]byte(value), &entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func FindPodEntity(id string, prefix string) (*corev1.Pod, error) {
	var entity *corev1.Pod

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return nil, fmt.Errorf("fail to find pod %s", err)
	}
	err = json.Unmarshal([]byte(value), &entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func FindHostNode(id string, prefix string) ([]*pb_common.InternalHostInfo, error) {
	var entity []*pb_common.InternalHostInfo

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return nil, fmt.Errorf("fail to find pod %s", err)
	}
	err = json.Unmarshal([]byte(value), &entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func FindComputenode(id string, prefix string) ([]*pb_common.InternalComputeInfo, error) {
	var entity []*pb_common.InternalComputeInfo

	value, err := Rdb.Get(Ctx, id+prefix).Result()
	if err != nil {
		return nil, fmt.Errorf("fail to find pod %s", err)
	}
	err = json.Unmarshal([]byte(value), &entity)
	if err != nil {
		return nil, err
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
		return entity, fmt.Errorf("fail to unmarsh value for key in DB %s", err)
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

		strippedKey := strings.Split(key, prefix)
		values[strippedKey[1]] = value
	}
	return values, nil
}
