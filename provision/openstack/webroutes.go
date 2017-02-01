package openstack

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

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
	fmt.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
	fmt.Println("Path", r.URL.Path)
	var path = strings.Trim(r.URL.Path, "/")
	s := strings.Split(path, "/")
	for i := 0; i < len(s); i++ {
		fmt.Printf("Path index %d equal %s\n", i, s[i])
	}

}

func canStartKetInstall() bool {
	return false
}

func startInstall() {

}
