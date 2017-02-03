package openstack

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type NodeMeta struct {
	Num int16
}

type AppContext struct {
	M map[string]NodeMeta
}

func ProvisionAndInstall(w http.ResponseWriter, r *http.Request) {
	log.Printf("ProvisionAndInstall called")
	//get the post data with credentials and setting
	var auth Auth
	//	var conf Config
	//	var nodeData serverData

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&auth)
	if err != nil {
		log.Printf("Error passing post data", err)
	}

	fmt.Println("Received", auth)
	defer r.Body.Close()
	//kick off all the requested nodes

	//	var _, err = buildNode(auth, conf, nodeData, "etcd")
	//	if err != nil {
	//		log.Println("Error instantiating Openstack client", err)
	//	}
}

func NodeUp(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.RawQuery)
	q, _ := url.ParseQuery(r.URL.RawQuery)
	//nodetype
	nodeType := q["type"][0]
	nodeIP := q["ip"][0]
	nodeName := q["nodeName"][0]
	fmt.Printf("Nodetype=%s nodeip=%s nodeName=%s", nodeType, nodeIP, nodeName)

}

func canStartKetInstall() bool {
	return false
}

func startInstall() {

}
