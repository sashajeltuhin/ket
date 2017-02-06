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

type NodesMeta struct {
	num  uint16
	name string
}

type CachedNodesMeta struct {
	Etcd   NodesMeta
	Master NodesMeta
	Worker NodesMeta
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

	//kick off all the requested nodes

	bag.Config.InstallscriptURL = ""

	for i := 0; i < int(bag.Opts.EtcdNodeCount); i++ {
		nodeName := buildFileName(bag.Opts.EtcdName, i)
		var _, erretcd = buildNode(bag.Auth, bag.Config, buildNodeData(nodeName, bag.Opts), bag.Opts, "etcd", ip)
		if erretcd != nil {
			log.Println("Error instantiating etcd node", erretcd)
		}
	}

	for i := 0; i < int(bag.Opts.MasterNodeCount); i++ {
		nodeName := buildFileName(bag.Opts.MasterName, i)
		var _, errMaster = buildNode(bag.Auth, bag.Config, buildNodeData(nodeName, bag.Opts), bag.Opts, "master", ip)
		if errMaster != nil {
			log.Println("Error instantiating master node", errMaster)
		}
	}

	for i := 0; i < int(bag.Opts.WorkerNodeCount); i++ {
		nodeName := buildFileName(bag.Opts.WorkerName, i)
		var _, errWorker = buildNode(bag.Auth, bag.Config, buildNodeData(nodeName, bag.Opts), bag.Opts, "worker", ip)
		if errWorker != nil {
			log.Println("Error instantiating worker node", errWorker)
		}
	}
}

func NodeUp(w http.ResponseWriter, r *http.Request) {
	log.Println("Node Up called")
	q := r.URL.Query()
	fmt.Println("received", q)
	nodeType := q["type"][0]
	nodeIP := q["ip"][0]
	nodeName := q["name"][0]
	log.Println("Parsed vals:", nodeType, nodeIP, nodeName)

	bag, bodyErr := parseBody(r)
	if bodyErr != nil {
		log.Println("Body error", bodyErr)
	}

	log.Println("Bag", bag)
	//save the provisioned node
	cacheNode(nodeName, nodeIP, bag)

	w.Header().Set("Content-Type", "application/json")
	resp := Response{Status: "Received node"}
	json.NewEncoder(w).Encode(resp)
	checkIfStartKetInstall(bag)
}

func cacheNode(nodeName string, nodeIP string, bag KetBag) {
	//sshpass -p passwd ssh-copy-id -i /ket/kismaticuser.key.pub -o StrictHostKeyChecking=no kismaticuser@nodeIP
	args := []string{"-p", bag.Opts.AdminPass, "ssh-copy-id", "-i", "/ket/kismaticuser.key.pub", "-o", "StrictHostKeyChecking=no", fmt.Sprintf("kismaticuser@%s", nodeIP)}
	log.Println("sshpass with args", args)
	out, err := exec.Command("sshpass", args...).Output()
	if err != nil {
		log.Println("Error pushing ssh cert", out, err)
	}

	log.Println("Caching new node:", nodeName)
	cached := KetNode{ID: nodeName, Host: nodeName, PrivateIPv4: nodeIP, PublicIPv4: nodeIP, SSHUser: bag.Opts.SSHUser}
	cachedJson, _ := json.Marshal(cached)
	fmt.Println("Marshaled node", string(cachedJson))
	errwrite := ioutil.WriteFile(nodeName, cachedJson, 0644)
	if errwrite != nil {
		log.Printf("Issues serializing cachedJson file %v\n", errwrite)
	}
}

func checkIfStartKetInstall(bag KetBag) {
	var check bool = false
	var nodeMeta CachedNodesMeta
	nodeMeta.Etcd = NodesMeta{num: bag.Opts.EtcdNodeCount, name: bag.Opts.EtcdName}
	nodeMeta.Master = NodesMeta{num: bag.Opts.MasterNodeCount, name: bag.Opts.MasterName}
	nodeMeta.Worker = NodesMeta{num: bag.Opts.WorkerNodeCount, name: bag.Opts.WorkerName}
	check = isCacheComplete(nodeMeta)
	if check {
		nodes := ProvisionedNodes{}
		for i := 0; i < int(nodeMeta.Etcd.num); i++ {
			fileName := buildFileName(nodeMeta.Etcd.name, i)
			n, cacheErr := getCachedNode(fileName)
			if cacheErr != nil {
				log.Println("Cannot read data for node", fileName, cacheErr)
			}
			nodes.Etcd = append(nodes.Etcd, KetNode{ID: n.Host, Host: n.Host, PublicIPv4: n.PublicIPv4, PrivateIPv4: n.PrivateIPv4, SSHUser: bag.Opts.SSHUser})
		}

		for i := 0; i < int(nodeMeta.Master.num); i++ {
			fileName := buildFileName(nodeMeta.Master.name, i)
			n, cacheErr := getCachedNode(fileName)
			if cacheErr != nil {
				log.Println("Cannot read data for node", fileName, cacheErr)
			}
			nodes.Master = append(nodes.Master, KetNode{ID: n.Host, Host: n.Host, PublicIPv4: n.PublicIPv4, PrivateIPv4: n.PrivateIPv4, SSHUser: bag.Opts.SSHUser})
		}

		for i := 0; i < int(nodeMeta.Worker.num); i++ {
			fileName := buildFileName(nodeMeta.Worker.name, i)
			n, cacheErr := getCachedNode(fileName)
			if cacheErr != nil {
				log.Println("Cannot read data for node", fileName, cacheErr)
			}
			nodes.Worker = append(nodes.Worker, KetNode{ID: n.Host, Host: n.Host, PublicIPv4: n.PublicIPv4, PrivateIPv4: n.PrivateIPv4, SSHUser: bag.Opts.SSHUser})
		}

		startInstall(bag.Opts, nodes)
	}
}

func buildFileName(fileName string, i int) string {
	if i > 0 {
		fileName = fmt.Sprintf("%s_%s", fileName, i)
	}
	return fileName
}

func startInstall(opts KetOpts, nodes ProvisionedNodes) {
	storageNodes := []KetNode{}
	if opts.Storage {
		storageNodes = []KetNode{nodes.Worker[0]}
	}
	fileName, err := makePlan(&Plan{
		AdminPassword:       opts.AdminPass,
		Etcd:                nodes.Etcd,
		Master:              nodes.Master,
		Worker:              nodes.Worker,
		Ingress:             []KetNode{nodes.Worker[0]},
		Storage:             storageNodes,
		MasterNodeFQDN:      nodes.Master[0].Host,
		MasterNodeShortName: nodes.Master[0].Host,
		SSHKeyFile:          opts.SSHFile,
		SSHUser:             opts.SSHUser,
	})
	if err != nil {
		log.Printf("Error creating plan", err)
		return
	}

	//"/ket/kismatic install apply -f " + fileName
	cmd := "/ket/kismatic"
	args := []string{"install", "apply", "-f", fileName}
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		log.Println("Error installing Kismatic", string(out), err)
	}
	log.Println("Kismatic Install:", string(out))

}

func makePlan(pln *Plan) (string, error) {
	template, err := template.New("planOverlay").Parse(OverlayNetworkPlan)
	if err != nil {
		return "", err
	}

	f, err := makeUniqueFile()
	if err != nil {
		return "", err
	}

	defer f.Close()
	w := bufio.NewWriter(f)

	if err = template.Execute(w, &pln); err != nil {
		return "", err
	}

	w.Flush()
	fmt.Println("To install your cluster, run:")
	fmt.Println("./kismatic install apply -f " + f.Name())

	return f.Name(), nil
}

func makeUniqueFile() (*os.File, error) {
	filename := "/ket/kismatic-cluster.yaml"

	return os.Create(filename)

}

func isCacheComplete(nodeMeta CachedNodesMeta) bool {
	//check etcd
	for i := 0; i < int(nodeMeta.Etcd.num); i++ {
		fileName := buildFileName(nodeMeta.Etcd.name, i)
		if doesFileExist(fileName) == false {
			return false
		}
	}
	for i := 0; i < int(nodeMeta.Master.num); i++ {
		fileName := buildFileName(nodeMeta.Master.name, i)
		if doesFileExist(fileName) == false {
			return false
		}
	}
	for i := 0; i < int(nodeMeta.Worker.num); i++ {
		fileName := buildFileName(nodeMeta.Worker.name, i)
		if doesFileExist(fileName) == false {
			return false
		}
	}

	log.Println("All nodes in place")
	return true
}

func doesFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil {
		log.Println("Node is not there yet", fileName)
		return false
	}
	log.Println("Node found", fileName)
	return true
}

func getCachedNode(fileName string) (KetNode, error) {
	var node KetNode
	dat, readerr := ioutil.ReadFile(fileName)
	fmt.Println("Opened node file", len(dat), readerr)
	if readerr == nil && dat != nil && len(dat) > 0 {
		deserr := json.Unmarshal(dat, &node)
		return node, deserr
	}

	return node, readerr

}
