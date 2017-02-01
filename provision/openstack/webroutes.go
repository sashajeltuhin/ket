package openstack

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func ProvisionAndInstall(w http.ResponseWriter, r *http.Request) {
	//get the post data with credentials and setting
	var auth Auth
	var conf Config
	var nodeData serverData
	//kick off all the requested nodes

	var _, err = buildNode(auth, conf, nodeData, "etcd")
	if err != nil {
		log.Println("Error instantiating Openstack client", err)
	}
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

	//post
	//	decoder := json.NewDecoder(req.Body)
	//	var t test_struct
	//	err := decoder.Decode(&t)
	//	if err != nil {
	//		panic(err)
	//	}
	//	defer req.Body.Close()
}

func canStartKetInstall() bool {
	return false
}

func startInstall() {

}
