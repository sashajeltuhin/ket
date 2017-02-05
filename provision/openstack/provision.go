package openstack

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type KetBag struct {
	Auth      Auth
	Config    Config
	Opts      KetOpts
	Installer KetNode
}

func GetClient(a Auth, conf Config) error {

	c := Client{}
	var err = c.getAPIClient(a, conf)
	fmt.Printf("GetClient called with error: %v\n", err)
	return err
}

func prepNodeTemplate(auth Auth, conf Config, nodeData serverData, opts KetOpts, nodeType string, webIP string) (map[string]string, error) {
	var tokens map[string]string = make(map[string]string)
	bag := KetBag{Auth: auth, Config: conf, Opts: opts, Installer: KetNode{Host: nodeData.Server.Name}}

	jsonStr, parseErr := json.Marshal(bag)
	if parseErr != nil {
		return nil, parseErr
	}
	tokens["nodeName"] = nodeData.Server.Name
	tokens["webPort"] = "8013"
	tokens["dcip"] = opts.DNSip
	tokens["domainName"] = opts.Domain
	tokens["domainSuf"] = opts.Suffix
	tokens["rootPass"] = opts.AdminPass
	tokens["nodeType"] = nodeType
	if webIP != "" {
		tokens["webIP"] = webIP
	}
	tokens["postData"] = b64.StdEncoding.EncodeToString([]byte(jsonStr))

	return tokens, nil
}

func buildNodeData(name string, opts KetOpts) serverData {
	var server serverData
	server.Server.Name = name
	server.Server.ImageRef = opts.Image
	server.Server.FlavorRef = opts.Flavor
	var n network
	n.Uuid = opts.Network
	server.Server.Networks = append(server.Server.Networks, n)
	var sec secgroup
	sec.Name = opts.SecGroup
	server.Server.Security_groups = append(server.Server.Security_groups, sec)
	return server
}

func buildNode(auth Auth, conf Config, nodeData serverData, opts KetOpts, nodeType string, webIP string) (string, error) {
	c := Client{}
	if conf.InstallscriptURL == "" {
		switch nodeType {
		case "install":
			conf.InstallscriptURL = "https://raw.githubusercontent.com/sashajeltuhin/ket/master/provision/openstack/scripts/ketinstall.sh"
			break
		default:
			conf.InstallscriptURL = "https://raw.githubusercontent.com/sashajeltuhin/ket/master/provision/openstack/scripts/ketnode.sh"
			break
		}
	}
	var script, scriptErr = c.downloadInitScript(conf.InstallscriptURL)
	if scriptErr != nil {
		return "", fmt.Errorf("Error downloading script %v", scriptErr)
	}
	scriptRaw := string(script)
	tokens, parseErr := prepNodeTemplate(auth, conf, nodeData, opts, nodeType, webIP)
	if parseErr != nil {
		return "", parseErr
	}
	for key := range tokens {
		tokenized := "^^" + key + "^^"
		scriptRaw = strings.Replace(scriptRaw, tokenized, tokens[key], 1)
	}
	nodeData.Server.User_data = b64.StdEncoding.EncodeToString([]byte(scriptRaw))
	var nodeID, err = c.buildNode(auth, conf, nodeData, nodeType)
	if err != nil {
		return "", fmt.Errorf("Error spinning up node %s. Error: %v", nodeData.Server.Name, err)
	}
	fmt.Printf("buildNode returned with error: %v\n", err)
	return nodeID, err
}
