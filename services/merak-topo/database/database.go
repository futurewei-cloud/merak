package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// {
//     "node":
//         {
//             "name": "a1"
//             "nics": [
//                 {
//                     "intf": "a1-intf1"
//                     "ip":"10.99.1.2"
//                 },
//                 {
//                     "intf": "a1-intf2"
//                     "ip":"10.99.1.3"
//                 }
//             ]

//         }
// }

type Nic struct {
	Intf string `json:"intf"`
	Ip   string `json:"ip"`
}

type Vnode struct {
	Name string `json:"name"`
	Nics []Nic  `json:"nics"`
}

// {
//     "link":
//        {
//             "name": "link1",
//             "local":
//                 {
//                     "name": "a2",
//                     "nics": "a2-intf1",
//                     "ip": ""
//                 },
//             "peer":
//                 {
//                     "name": "a3",
//                     "nics": "a3-intf1",
//                     "ip": ""
//                 }
//         }
// }

type Vport struct {
	Name string `json:"name"`
	Nics string `json:"nics"`
	Ip   string `json:"ip"`
}

type Vlink struct {
	Name  string `json:"name"`
	Local Vport  `json:"local"`
	Peer  Vport  `json:"peer"`
}

// {
//     "topology_id": "",
//     "vnodes": ["a1","a2","a3"],
//     "vlinks": ["link1"]
// }

type TopologyData struct {
	Topology_id string  `json:"topology_id"`
	Vnodes      []Vnode `json:"vnodes"`
	Vlinks      []Vlink `json:"vlinks"`
}

// var Topo TopologyData
// var Node Vnode
// var Link Vlink

// func Topo2Json(d *TopologyData) []byte {
// 	j, err := json.Marshal(d)
// 	if err != nil {
// 		fmt.Println("error:", err)
// 	}
// 	return j
// }

// func Node2Json(n *Vnode) []byte {
// 	j, err := json.Marshal(n)
// 	if err != nil {
// 		fmt.Println("error:", err)
// 	}
// 	return j
// }

// func Link2Json(l *Vlink) []byte {
// 	j, err := json.Marshal(l)
// 	if err != nil {
// 		fmt.Println("error:", err)
// 	}
// 	return j
// }

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
