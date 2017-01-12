package vagrant

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/apprenda/kismatic-provision/provision/utils"
)

type NodeType uint32

const (
	Etcd NodeType = 1 << iota
	Master
	Worker
	Ingress
)

var NodeTypes = []NodeType{Etcd, Master, Worker, Ingress}

var NodeTypeStrings = map[NodeType]string{
	Etcd:    "etcd",
	Master:  "master",
	Worker:  "worker",
	Ingress: "ingress",
}

type InfrastructureOpts struct {
	Count             map[NodeType]uint16
	OverlapRoles      bool
	NodeCIDR          string
	Redhat            bool
	PrivateSSHKeyPath string
	Vagrantfile       string
}

type NodeDetails struct {
	Name  string
	IP    net.IP
	Types NodeType
}

type Infrastructure struct {
	Network           net.IPNet
	Broadcast         net.IP
	Nodes             []NodeDetails
	DNSReflector      string
	PrivateSSHKeyPath string
	PublicSSHKeyPath  string
}

func NewInfrastructure(opts *InfrastructureOpts) (*Infrastructure, error) {
	_, network, err := net.ParseCIDR(opts.NodeCIDR)

	if err != nil {
		return nil, err
	}

	broadcast, err := utils.BroadcastIPv4(*network)
	if err != nil {
		return nil, err
	}

	i := &Infrastructure{
		Network:   *network,
		Broadcast: broadcast,
		Nodes:     []NodeDetails{},
	}

	sshError := i.ensureSSHKeys(opts.PrivateSSHKeyPath)
	if sshError != nil {
		return nil, sshError
	}

	var overlapTypes NodeType

	// keep creating nodes until counts are exhausted
	for j := uint16(1); ; j++ {
		overlapTypes = NodeType(0)
		finished := true

		for _, nodeType := range NodeTypes {
			if j <= opts.Count[nodeType] {
				if opts.OverlapRoles {
					overlapTypes |= nodeType
				} else {
					_, err := i.appendNode(j, NodeTypeStrings[nodeType], nodeType)
					if err != nil {
						return i, err
					}
				}
				finished = false
			}
		}

		if overlapTypes > 0 {
			_, err := i.appendNode(j, "node", overlapTypes)
			if err != nil {
				return i, err
			}
		}

		if finished {
			break
		}
	}

	return i, nil
}

func (i *Infrastructure) ensureSSHKeys(privateSSHKeyPath string) error {

	if privateSSHKeyPath == "" {
		i.PrivateSSHKeyPath = "kismatic-cluster.pem"
	} else {
		i.PrivateSSHKeyPath = privateSSHKeyPath
	}

	// ensure absolute path
	var absErr error
	i.PrivateSSHKeyPath, absErr = filepath.Abs(i.PrivateSSHKeyPath)
	if absErr != nil {
		return absErr
	}

	i.PublicSSHKeyPath = i.PrivateSSHKeyPath + ".pub"

	privateKey, privateKeyErr := utils.LoadOrCreatePrivateSSHKey(i.PrivateSSHKeyPath)
	if privateKeyErr != nil {
		return privateKeyErr
	}

	publicKeyErr := utils.CreatePublicKey(privateKey, i.PublicSSHKeyPath)
	if publicKeyErr != nil {
		return publicKeyErr
	}

	// ensure correct permissions
	os.Chmod(i.PrivateSSHKeyPath, 0600)
	os.Chmod(i.PublicSSHKeyPath, 0600)

	return nil
}

func (i *Infrastructure) appendNode(nodeIndex uint16, name string, types NodeType) (*NodeDetails, error) {
	ip, err := i.nextNodeIP()

	if err != nil {
		return nil, err
	}

	hostname := fmt.Sprintf("%v%03d", name, nodeIndex)

	node := NodeDetails{
		Name:  hostname,
		IP:    ip,
		Types: types,
	}

	i.Nodes = append(i.Nodes, node)

	return &node, nil
}

func (i *Infrastructure) nextNodeIP() (net.IP, error) {
	var ip net.IP
	var err error

	if len(i.Nodes) < 1 {
		ip = i.Network.IP

		// increment by 1 to account for gateway
		ip, err = utils.IncrementIPv4(ip)
		if err != nil {
			return nil, err
		}
	} else {
		lastNode := i.Nodes[len(i.Nodes)-1:][0]
		ip = lastNode.IP
	}

	ip, err = utils.IncrementIPv4(ip)
	if err != nil {
		return nil, err
	}

	// assumes broadcast address is last host in CIDR range
	if !i.Network.Contains(ip) || i.Broadcast.Equal(ip) {
		return ip, errors.New("infrastructure: ip address overflowed available cidr range")
	}

	return ip, nil
}

func (i *Infrastructure) nodesByType(nodeType NodeType) []NodeDetails {
	filtered := []NodeDetails{}
	for _, node := range i.Nodes {
		if (node.Types & nodeType) > NodeType(0) {
			filtered = append(filtered, node)
		}
	}
	return filtered
}
