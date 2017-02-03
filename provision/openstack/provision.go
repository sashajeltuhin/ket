package openstack

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type KetBag struct {
	Auth   Auth
	Config Config
	Opts   KetOpts
}

func GetClient(a Auth, conf Config) error {

	c := Client{}
	var err = c.getAPIClient(a, conf)
	fmt.Printf("GetClient called with error: %v\n", err)
	return err
}

func prepNodeTemplate(auth Auth, conf Config, nodeData serverData, opts KetOpts, nodeType string) (map[string]string, error) {
	var tokens map[string]string = make(map[string]string)
	bag := KetBag{Auth: auth, Config: conf, Opts: opts}
	switch nodeType {
	case "install":
		jsonStr, parseErr := json.Marshal(bag)
		if parseErr != nil {
			return nil, parseErr
		}
		fmt.Printf("Auth formatted: %v\n", string(jsonStr))
		tokens["nodeName"] = nodeData.Server.Name
		tokens["webPort"] = "8013"
		tokens["dcip"] = opts.DNSip
		tokens["domainName"] = opts.Domain
		tokens["domainSuf"] = opts.Suffix
		tokens["rootPass"] = opts.AdminPass
		tokens["postData"] = string(jsonStr)
		break
	}

	return tokens, nil
}

func buildNode(auth Auth, conf Config, nodeData serverData, opts KetOpts, nodeType string) (string, error) {
	c := Client{}
	if conf.InstallscriptURL == "" {
		switch nodeType {
		case "install":
			conf.InstallscriptURL = "https://raw.githubusercontent.com/sashajeltuhin/ket/master/provision/openstack/scripts/ketinstall.sh"
			break
		}
	}
	var script, scriptErr = c.downloadInitScript(conf.InstallscriptURL)
	if scriptErr != nil {
		return "", fmt.Errorf("Error downloading script %v", scriptErr)
	}
	scriptRaw := string(script)
	tokens, parseErr := prepNodeTemplate(auth, conf, nodeData, opts, nodeType)
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
