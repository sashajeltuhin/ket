package openstack

import (
	"bufio"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
)

type ProvisionedNodes struct {
	Etcd   []KetNode
	Master []KetNode
	Worker []KetNode
}

type Response struct {
	Status string `json:"status"`
}

func parseBody(r *http.Request) (KetBag, error) {
	//get the post data with credentials and setting
	bag := KetBag{}
	//	var conf Config
	//	var nodeData serverData
	bodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body", err)
	}

	decoded, decerr := b64.StdEncoding.DecodeString(string(bodyData))
	if decerr != nil {
		log.Println("Something wrong with the post data 64bit encoding", decerr)
		return bag, decerr
	}

	unmarshErr := json.Unmarshal(decoded, &bag)
	if unmarshErr != nil {
		log.Println("Cannot deserialize ket bag", unmarshErr)
		return bag, unmarshErr
	}
	fmt.Println("Bag", bag)
	defer r.Body.Close()
	return bag, nil
}

func ProvisionAndInstall(w http.ResponseWriter, r *http.Request) {
	log.Printf("ProvisionAndInstall called")
	log.Println(r.URL.RawQuery)
	q, _ := url.ParseQuery(r.URL.RawQuery)
	var ip string = "10.20.50.199"
	if len(q["ip"]) > 0 {
		ip = q["ip"][0]
	}
	bag, bodyErr := parseBody(r)
	if bodyErr != nil {
		log.Println("Body error", bodyErr)
	}

	addToDNS(bag.Opts.DNSip, bag.Installer.Host, bag.Opts.Domain, bag.Opts.Suffix, ip)
	//kick off all the requested nodes

	var _, erretcd = buildNode(bag.Auth, bag.Config, buildNodeData("ketautoetcd", bag.Opts), bag.Opts, "etcd", ip)
	if erretcd != nil {
		log.Println("Error instantiating etcd node", erretcd)
	}
	var _, errMaster = buildNode(bag.Auth, bag.Config, buildNodeData("ketautomaster", bag.Opts), bag.Opts, "master", ip)
	if errMaster != nil {
		log.Println("Error instantiating master node", errMaster)
	}
	var _, errWorker = buildNode(bag.Auth, bag.Config, buildNodeData("ketautoworker", bag.Opts), bag.Opts, "worker", ip)
	if errWorker != nil {
		log.Println("Error instantiating worker node", errWorker)
	}
}

func NodeUp(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fmt.Println("received", q)
	nodeType := q["type"][0]
	nodeIP := q["ip"][0]
	nodeName := q["name"][0]
	fmt.Println("Parsed vals:", nodeType, nodeIP, nodeName)

	bag, bodyErr := parseBody(r)
	if bodyErr != nil {
		log.Println("Body error", bodyErr)
	}
	addToDNS(bag.Opts.DNSip, bag.Installer.Host, bag.Opts.Domain, bag.Opts.Suffix, nodeIP)

	w.Header().Set("Content-Type", "application/json")
	resp := Response{Status: "Received node"}
	json.NewEncoder(w).Encode(resp)
	checkIfStartKetInstall(bag)
}

func checkIfStartKetInstall(bag KetBag) {
	var check bool = false
	if check {
		nodes := ProvisionedNodes{}
		nodes.Etcd = append(nodes.Etcd, KetNode{ID: "1", Host: "ketautoetcd", PublicIPv4: "10.20.50.1", PrivateIPv4: "10.20.50.1", SSHUser: bag.Opts.SSHUser})
		nodes.Master = append(nodes.Master, KetNode{ID: "1", Host: "ketautomaster", PublicIPv4: "10.20.50.1", PrivateIPv4: "10.20.50.1", SSHUser: bag.Opts.SSHUser})
		nodes.Worker = append(nodes.Worker, KetNode{ID: "1", Host: "ketautoworker", PublicIPv4: "10.20.50.1", PrivateIPv4: "10.20.50.1", SSHUser: bag.Opts.SSHUser})
		startInstall(bag.Opts, nodes)
	}
}

func startInstall(opts KetOpts, nodes ProvisionedNodes) {
	storageNodes := []KetNode{}
	if opts.Storage {
		storageNodes = []KetNode{nodes.Worker[0]}
	}
	err := makePlan(&Plan{
		AdminPassword:       opts.AdminPass,
		Etcd:                nodes.Etcd,
		Master:              nodes.Master,
		Worker:              nodes.Worker,
		Ingress:             []KetNode{nodes.Worker[0]},
		Storage:             storageNodes,
		MasterNodeFQDN:      nodes.Master[0].PublicIPv4,
		MasterNodeShortName: nodes.Master[0].PrivateIPv4,
		SSHKeyFile:          opts.SSHFile,
		SSHUser:             opts.SSHUser,
	})
	if err != nil {
		log.Printf("Error creating plan", err)
	}
}

func addToDNS(dns string, serverName string, domain string, suf string, ip string) {
	cmd := "echo -e \"server " + dns + "\\nupdate add " + serverName + "." + domain + "." + suf + " 3600 A " + ip + "\\nsend\\n\" | nsupdate -v"
	log.Println("Exec command dns:", cmd)
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		log.Println("Error adding to DNS", out, err)
	}
	log.Println("Add dns:", out)
	//echo -e "server 10.0.0.1\nupdate add host.domain.nl 3600 A 10.0.0.2\nsend\n" | nsupdate -v
}

func makePlan(pln *Plan) error {
	template, err := template.New("planOverlay").Parse(OverlayNetworkPlan)
	if err != nil {
		return err
	}

	f, err := makeUniqueFile()
	if err != nil {
		return err
	}

	defer f.Close()
	w := bufio.NewWriter(f)

	if err = template.Execute(w, &pln); err != nil {
		return err
	}

	w.Flush()
	fmt.Println("To install your cluster, run:")
	fmt.Println("./kismatic install apply -f " + f.Name())

	return nil
}

func makeUniqueFile() (*os.File, error) {
	filename := "/ket/kismatic-cluster.yaml"

	return os.Create(filename)

}
