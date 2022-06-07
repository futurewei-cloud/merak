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
//             "pairs":
//                 {
//                     "local_name": "a2"
//                     "local_nics": "a2-intf1"
//                     "local_ip": ""
//                     "peer_name": "a3"
//                     "peer_nics": "a3-intf1"
//                     "peer_ip": ""
//                 }

//         }

// }

// type Pair struct {
// 	Local_name string `json:"local_name"`
// 	Local_nics string `json:"local_nics"`
// 	Local_ip   string `json:"local_ip"`
// 	Peer_name  string `json:"peer_name"`
// 	Peer_nics  string `json:"peer_nics"`
// 	Peer_ip    string `json:"peer_ip"`
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
	Topology_id string   `json:"topology_id"`
	Vnodes      []string `json:"vnodes"`
	Vlinks      []string `json:"vlinks"`
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

func CreateDBClient() {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	RDB = rdb
}

func SetValue(key string, val interface{}) {
	j, err := json.Marshal(val)
	if err != nil {
		fmt.Println("error:", err)
	}
	err2 := RDB.Set(Ctx, key, j, 0).Err()
	if err2 != nil {
		panic(err2)
	} else {
		fmt.Printf("Save data in Redis")
	}
}

func GetValue(key string) string {
	val, err := RDB.Get(Ctx, key).Result()
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("Get data in Redis")
	}
	return val
}
