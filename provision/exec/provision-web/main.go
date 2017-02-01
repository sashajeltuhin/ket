package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sashajeltuhin/ket/provision/openstack"
)

func main() {
	fmt.Println("Listening on port 8013")
	http.HandleFunc("/nodeup/etcd", openstack.NodeUp)
	http.HandleFunc("/nodeup/etcd", openstack.ProvisionAndInstall)
	err := http.ListenAndServe(":8013", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
