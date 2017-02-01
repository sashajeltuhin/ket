package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/apprenda/kismatic-provision/provision/openstack"
)

func main() {
	fmt.Println("Listening on port 8013")
	http.HandleFunc("/nodeup/etcd", openstack.NodeUp)
	err := http.ListenAndServe(":8013", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
