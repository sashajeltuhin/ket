package openstack

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

func GetClient(a Auth, conf Config) error {

	c := Client{}
	var err = c.getAPIClient(a, conf)
	fmt.Printf("GetClient called with error: %v\n", err)
	return err
}

func prepNodeTemplate(auth Auth, conf Config, nodeType string) (map[string]string, error) {
	var tokens map[string]string = make(map[string]string)
	switch nodeType {
	case "install":
		jsonStr, parseErr := json.Marshal(auth)
		if parseErr != nil {
			return nil, parseErr
		}
		fmt.Printf("Auth formatted: %v\n", string(jsonStr))
		tokens["webPort"] = "8013"
		tokens["nodeName"] = "ketinstall"
		tokens["rootPass"] = "@ppr3nda"
		tokens["postData"] = string(jsonStr)
		break
	}

	return tokens, nil
}

func buildNode(auth Auth, conf Config, nodeData serverData, nodeType string) (string, error) {
	c := Client{}
	if conf.installscriptURL == "" {
		switch nodeType {
		case "install":
			conf.installscriptURL = "https://raw.githubusercontent.com/sashajeltuhin/ket/master/provision/openstack/scripts/ketinstall.sh"
			break
		}
	}
	var script, scriptErr = c.downloadInitScript(conf.installscriptURL)
	if scriptErr != nil {
		return "", fmt.Errorf("Error downloading script %v", scriptErr)
	}
	scriptRaw := string(script)
	tokens, parseErr := prepNodeTemplate(auth, conf, nodeType)
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
