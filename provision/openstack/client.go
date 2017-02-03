package openstack

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	s "strings"
	"time"

	"github.com/Jeffail/gabs"
)

type Config struct {
	Urlauth          string
	Apiverauth       string
	Urlcomp          string
	Apivercomp       string
	InstallscriptURL string
}

// Credentials to be used for accessing the AI
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Auth struct {
	Body struct {
		Credentials Credentials `json:"passwordCredentials"`
		Tenant      string      `json:"tenantId"`
	} `json:"auth"`
}

// Client for provisioning machines on Openstack
type Client struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

type serverData struct {
	Server struct {
		Name            string     `json:"name"`
		User_data       string     `json:"user_data"`
		Key_name        string     `json:"key_name,omitempty"`
		ImageRef        string     `json:"imageRef"`
		FlavorRef       string     `json:"flavorRef"`
		Networks        []network  `json:"networks"`
		Security_groups []secgroup `json:"security_groups"`
	} `json:"server"`
}

type network struct {
	Uuid string `json:"uuid"`
}

type secgroup struct {
	Name string `json:"name"`
}

func isValidJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func (c *Client) login(a Auth, conf Config) (string, error) {

	var fileName = a.Body.Credentials.Username + ".token"
	dat, readerr := ioutil.ReadFile(fileName)
	if readerr != nil && dat != nil && len(dat) > 0 {
		fmt.Printf("Opened token file with content %s", string(dat))
		var savedCreds Client
		deserr := json.Unmarshal(dat, savedCreds)
		if deserr == nil {
			now := time.Now()
			exptime := savedCreds.Expires
			fmt.Println("About to use time", exptime)
			if now.Before(exptime) {
				fmt.Println("Will use old token", c.Token)
				return c.Token, nil
			}
		}
	}

	fmt.Errorf("Token file %s does not exist ", fileName)

	url := conf.Urlauth + conf.Apiverauth + "/tokens"
	fmt.Println("Openstack URL:>", url)

	jsonStr, parseErr := json.Marshal(a)
	if parseErr != nil {
		fmt.Errorf("Something is wrong with auth body", parseErr)
		return "", fmt.Errorf("Something is wrong with auth body: %v", parseErr)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	tr := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: &tr,
		Timeout:   30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	body, _ := ioutil.ReadAll(resp.Body)

	jsonParsed, err := gabs.ParseJSON(body)

	c.Token = jsonParsed.Path("access.token.id").String()
	c.Token = s.Trim(c.Token, "\"")
	var exp string = jsonParsed.Path("access.token.expires").String()
	exp = s.Trim(exp, "\"")
	fmt.Println("Expiration", exp)
	layout := "2006-01-02T15:04:05Z"
	c.Expires, _ = time.Parse(layout, exp)
	fmt.Println("Time object", c.Expires)
	fmt.Println("login results:", c.Token)
	credsJson, _ := json.Marshal(*c)
	fmt.Println("Marshaled creds", string(credsJson))
	errwrite := ioutil.WriteFile(fileName, credsJson, 0644)
	if errwrite != nil {
		fmt.Errorf("Issues serializing token file %v\n", errwrite)
	}

	return c.Token, err
}

func (c *Client) getAPIClient(auth Auth, conf Config) error {
	if c.Token == "" {
		token, err := c.login(auth, conf)
		if err != nil {
			return fmt.Errorf("Error with credentials provided: %v", err)
		}
		fmt.Printf("Returned token %s \n", token)

	}
	return nil
}

func (c *Client) buildNode(auth Auth, conf Config, nodeData serverData, nodeType string) (string, error) {
	token, err := c.login(auth, conf)
	if err != nil {
		return "", fmt.Errorf("Error with auth: %v", err)
	}

	var url = conf.Urlcomp + conf.Apivercomp + "/" + auth.Body.Tenant + "/servers"

	fmt.Println("After login returned with token", token)

	jsonStr, parseErr := json.Marshal(nodeData)
	if parseErr != nil {
		return "", fmt.Errorf("Error with server data format: %v", parseErr)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("Content-Type", "application/json")

	tr := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: &tr,
		Timeout:   30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	body, _ := ioutil.ReadAll(resp.Body)

	if isValidJSON(string(body)) == false {
		return "", errors.New("Not a valid JSON response: " + string(body))
	}

	jsonParsed, err := gabs.ParseJSON(body)

	var message, nodeID string

	if jsonParsed.Exists("overLimit", "code") {
		var code = jsonParsed.Path("overLimit.code").String()
		if code == "413" {
			message = "Out of space"
			fmt.Println("message:", message)
		}
	}

	if jsonParsed.Exists("badRequest") {
		message = jsonParsed.Path("badRequest.message").String()
		fmt.Println("message:", message)

	}
	if jsonParsed.Exists("server", "id") {
		nodeID = jsonParsed.Path("server.id").String()
		fmt.Println("nodeID:", nodeID)
	}
	if nodeID == "" {
		message = "Unknown error"
	}
	if message != "" {
		return "", errors.New(message)
	}

	return nodeID, nil

}

func (c *Client) downloadInitScript(url string) (string, error) {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	return string(body), nil

}
