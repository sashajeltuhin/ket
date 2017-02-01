package openstack

import (
	b64 "encoding/base64"
	"fmt"
)

func GetClient(a Auth, conf Config) error {

	c := Client{}
	var err = c.getAPIClient(a, conf)
	fmt.Printf("GetClient called with error: %v\n", err)
	return err
}

func buildNode(auth Auth, conf Config, nodeData serverData, nodeType string) (string, error) {
	c := Client{}
	var script, scriptErr = c.downloadInitScript(nodeType, "")
	if scriptErr != nil {
		return "", fmt.Errorf("Error downloading script %v", scriptErr)
	}
	nodeData.Server.User_data = b64.StdEncoding.EncodeToString(script)
	var nodeID, err = c.buildNode(auth, conf, nodeData, nodeType)
	if err != nil {
		return "", fmt.Errorf("Error spinning up node %s. Error: %v", nodeData.Server.Name, err)
	}
	fmt.Printf("buildNode returned with error: %v\n", err)
	return nodeID, err
}
