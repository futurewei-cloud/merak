package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"log"
	"net"
	"time"
)

type message struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	ID     int           `json:"id"`
}

// ========================================================================================
type columns struct {
	Columns []string `json:"columns"`
}
type id2 struct {
	Port        columns `json:"Port"`
	Interface   columns `json:"Interface"`
	Controller  columns `json:"Controller"`
	Bridge      columns `json:"Bridge"`
	OpenVSwitch columns `json:"Open_vSwitch"`
}

// ========================================================================================

type interfaceInsertRow struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
type interfaceInsert struct {
	UuidName string             `json:"uuid-name"`
	Row      interfaceInsertRow `json:"row"`
	Op       string             `json:"op"`
	Table    string             `json:"table"`
}

// ========================================
type portInsertRow struct {
	Name       string   `json:"name"`
	Interfaces []string `json:"interfaces"`
}
type portInsert struct {
	UuidName string        `json:"uuid-name"`
	Row      portInsertRow `json:"row"`
	Op       string        `json:"op"`
	Table    string        `json:"table"`
}

// ========================================

type bridgeUpdateRow struct {
	Ports []interface{} `json:"ports"`
}
type bridgeUpdate struct {
	Where [][]interface{} `json:"where"`
	Row   bridgeUpdateRow `json:"row"`
	Op    string          `json:"op"`
	Table string          `json:"table"`
}

// ========================================

type ovsMutate struct {
	Mutations [][]interface{} `json:"mutations"`
	Where     [][]interface{} `json:"where"`
	Op        string          `json:"op"`
	Table     string          `json:"table"`
}

// ========================================
type ovsSelect struct {
	Where   [][]interface{} `json:"where"`
	Columns []string        `json:"columns"`
	Op      string          `json:"op"`
	Table   string          `json:"table"`
}

// ========================================
type comment struct {
	Comment string `json:"comment"`
	Op      string `json:"op"`
}

// ========================================================================================

type id2ReturnInterfaceIdInitial struct {
	Name   string `json:"name"`
	Ofport int    `json:"ofport"`
	Type   string `json:"type"`
}
type id2ReturnInterfaceId struct {
	Initial id2ReturnInterfaceIdInitial `json:"initial"`
}
type id2ReturnInterfaceIds map[string]id2ReturnInterfaceId

type id2ReturnPortIdInitial struct {
	Name       string   `json:"name"`
	Interfaces []string `json:"interfaces"`
}
type id2ReturnPortId struct {
	Initial id2ReturnPortIdInitial `json:"initial"`
}
type id2ReturnPortIds map[string]id2ReturnPortId

type id2ReturnBridgeIdInitial struct {
	Name  string   `json:"name"`
	Ports []string `json:"ports"`
}
type id2ReturnBridgeId struct {
	Initial id2ReturnBridgeIdInitial `json:"initial"`
}
type id2ReturnBridgeIds map[string]id2ReturnBridgeId

type id2ReturnOvsIdInitial struct {
	Bridges []string `json:"bridges"`
	CurCfg  int      `json:"cur_cfg"`
}
type id2ReturnOvsId struct {
	Initial id2ReturnOvsIdInitial `json:"initial"`
}
type id2ReturnOvsIds map[string]id2ReturnOvsId

type id2Return struct {
	Id     int `json:"id"`
	Result struct {
		Interface   id2ReturnInterfaceIds `json:"Interface"`
		Port        id2ReturnPortIds      `json:"Port"`
		Bridge      id2ReturnBridgeIds    `json:"Bridge"`
		OpenVSwitch id2ReturnOvsIds       `json:"Open_vSwitch"`
	} `json:"result"`
	Error interface{} `json:"error"`
}

// ========================================================================================
var id2ReturnJson id2Return

// ========================================================================================
func reader(r io.Reader) {
	buf := make([]byte, 1024)
	//for {
	n, err := r.Read(buf[:])
	if err != nil {
		return
	}
	//println("Client got:", string(buf[0:n]))
	json.Unmarshal(buf[0:n], &id2ReturnJson)
	log.Printf("Client got: %+v", id2ReturnJson)
	//}
}

// ========================================================================================

func main() {
	c, err := net.Dial("unix", "/var/run/openvswitch/db.sock")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	go reader(c)
	id2Call := &message{
		Method: "monitor_cond",
		Params: []interface{}{"Open_vSwitch", []string{"monid", "Open_vSwitch"}, id2{
			Port:        columns{Columns: []string{"fake_bridge", "interfaces", "name", "tag"}},
			Interface:   columns{Columns: []string{"error", "name", "ofport", "type"}},
			Controller:  columns{Columns: []string{}},
			Bridge:      columns{Columns: []string{"controller", "fail_mode", "name", "ports"}},
			OpenVSwitch: columns{Columns: []string{"bridges", "cur_cfg"}},
		}},
		ID: 2,
	}
	id2CallJson, err := json.Marshal(id2Call)
	//for {
	_, write_err := c.Write([]byte(id2CallJson))
	if write_err != nil {
		log.Fatal("write error:", write_err)
		//break
	}

	var bridgeId string
	var bridgePort string
	var ovsId string
	time.Sleep(2 * time.Second)
	for k, _ := range id2ReturnJson.Result.OpenVSwitch {
		// it will only loop once
		ovsId = k
		//log.Println("ovsId: ", k)
	}
	for k, v := range id2ReturnJson.Result.Bridge {
		// it will only loop once
		bridgeId = k
		bridgePort = v.Initial.Ports[1]
		//log.Println("ovsId: ", k)
	}
	//bridgePort = id2ReturnJson.Result.Bridge[0].Initial.Ports[1]

	log.Println("ovsId: ", ovsId)
	log.Println("bridgeId: ", bridgeId)
	log.Println("bridgePort", bridgePort)

	id4Call := &message{
		Method: "transact",
		Params: []interface{}{
			"Open_vSwitch",
			ovsMutate{
				Mutations: [][]interface{}{{"next_cfg", "+=", 1}},
				Where:     [][]interface{}{{"_uuid", "==", []string{"uuid", ovsId}}},
				Op:        "mutate",
				Table:     "Open_vSwitch",
			},
			ovsSelect{
				Where:   [][]interface{}{{"_uuid", "==", []string{"uuid", ovsId}}},
				Columns: []string{"next_cfg"},
				Op:      "select",
				Table:   "Open_vSwitch",
			},
			comment{
				Comment: "ovs-vsctl (invoked by -bash): ovs-vsctl -v add-port br-int try -- set Interface try type=internal",
				Op:      "comment",
			},
		},
		ID: 4,
	}

	bridgeUpdateNew := &bridgeUpdate{
		Where: [][]interface{}{{"_uuid", "==", []string{"uuid", bridgeId}}},
		Row: bridgeUpdateRow{Ports: []interface{}{"set", []interface{}{
			[]string{"uuid", bridgePort},
		}}},
		Op:    "update",
		Table: "Bridge",
	}

	for i := 1; i < 5; i++ {
		portUUID := uuid.New().String()
		interfaceUUID := uuid.New().String()
		newName := "try" + string(i)
		interfaceNew := &interfaceInsert{
			UuidName: interfaceUUID,
			Row:      interfaceInsertRow{Name: newName, Type: "internal"},
			Op:       "insert",
			Table:    "Interface",
		}
		id4Call.Params = append(id4Call.Params, interfaceNew)

		portNew := &portInsert{
			UuidName: portUUID,
			Row:      portInsertRow{Name: newName, Interfaces: []string{"named-uuid", interfaceUUID}},
			Op:       "insert",
			Table:    "Port",
		}
		id4Call.Params = append(id4Call.Params, portNew)

		bridgeUpdateNew.Row.Ports = append(id4Call.Params, []string{"named-uuid", portUUID})
	}
	id4Call.Params = append(id4Call.Params, bridgeUpdateNew)

	id4Call_json, err := json.Marshal(id4Call)
	//for {
	_, id4Call_err := c.Write([]byte(id4Call_json))
	if id4Call_err != nil {
		log.Fatal("write error:", write_err)
		//break
	}
	time.Sleep(1e9)

	//}
}
