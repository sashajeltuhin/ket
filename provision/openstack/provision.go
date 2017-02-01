package openstack

import (
	b64 "encoding/base64"
	"fmt"
	"strings"
)

func GetClient(a Auth, conf Config) error {

	c := Client{}
	var err = c.getAPIClient(a, conf)
	fmt.Printf("GetClient called with error: %v\n", err)
	return err
}

func prepNodeTemplate(auth Auth, conf Config, nodeType string) map[string]string {
	var tokens map[string]string = make(map[string]string)
	switch nodeType {
	case "install":
		tokens["webPort"] = "8013"
		tokens["nodeName"] = "ketinstall"
		tokens["rootPass"] = "@ppr3nda"
		break
	}

	return tokens
}

func buildNode(auth Auth, conf Config, nodeData serverData, nodeType string) (string, error) {
	c := Client{}
	if conf.installscriptURL == "" {
		switch nodeType {
		case "install":
			conf.installscriptURL = "http://installs.apprendalabs.com/installscripts/ketinstall.sh"
			break
		}
	}
	var script, scriptErr = c.downloadInitScript(conf.installscriptURL)
	if scriptErr != nil {
		return "", fmt.Errorf("Error downloading script %v", scriptErr)
	}
	scriptRaw := string(script)
	tokens := prepNodeTemplate(auth, conf, nodeType)
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
