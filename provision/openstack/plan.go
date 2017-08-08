package openstack

type KetNode struct {
	ID          string
	Host        string
	PublicIPv4  string
	PrivateIPv4 string
	SSHUser     string
}

type Plan struct {
	Etcd                []KetNode
	Master              []KetNode
	Worker              []KetNode
	Ingress             []KetNode
	Storage             []KetNode
	MasterNodeFQDN      string
	MasterNodeShortName string
	SSHUser             string
	SSHKeyFile          string
	AdminPassword       string
	CNI                 string
}

const OverlayNetworkPlan = `cluster:
  name: kubernetes
  admin_password: "{{.AdminPassword}}"      # This password is used to login to the Kubernetes Dashboard and can also be used for administration without a security certificate
  disable_package_installation: false    # When true, installation will not occur if any node is missing the correct deb/rpm packages. When false, the installer will attempt to install missing packages for you.
  package_repository_urls: ""            # Comma-separated list of URLs of the repositories that should be used during installation. These repositories must contain the kismatic packages and all their transitive dependencies.
  disconnected_installation: false       # Set to true if you have already installed the required packages on the nodes or provided a local URL in package_repository_urls containing those packages.
  disable_registry_seeding: false        # Set to true if you have seeded your registry with the required images for the installation.
  networking:
    pod_cidr_block: 172.16.0.0/16        # Kubernetes will assign pods IPs in this range. Do not use a range that is already in use on your local network!
    service_cidr_block: 172.20.0.0/16    # Kubernetes will assign services IPs in this range. Do not use a range that is already in use by your local network or pod network!
    update_hosts_files: false            # When true, the installer will add entries for all nodes to other nodes' hosts files. Use when you don't have access to DNS.
    http_proxy: ""                       # Set the proxy server to use for HTTP connections.
    https_proxy: ""                      # Set the proxy server to use for HTTPs connections
    no_proxy: ""                         # List of host names and/or IPs that shouldn't go through any proxy. If set to a asterisk '*' only, it matches all hosts.
  certificates:
    expiry: 17520h                       # Self-signed certificate expiration period in hours; default is 2 years.
    ca_expiry: 17520h                    # CA certificate expiration period in hours; default is 2 years.
  ssh:
    user: {{.SSHUser}}
    ssh_key: {{.SSHKeyFile}}             # Absolute path to the ssh public key we should use to manage nodes.
    ssh_port: 22
  kube_apiserver:
    option_overrides: {}
docker:
  storage:
    direct_lvm:                          # Configure devicemapper in direct-lvm mode (RHEL/CentOS only).
      enabled: false
      block_device: ""                   # Path to the block device that will be used for direct-lvm mode. This device will be wiped and used exclusively by docker.
      enable_deferred_deletion: false    # Set to true if you want to enable deferred deletion when using direct-lvm mode.
docker_registry:                         # Here you will provide the details of your Docker registry or setup an internal one to run in the cluster. This is optional and the cluster will always have access to the Docker Hub.
  setup_internal: true                  # When true, a Docker Registry will be installed on top of your cluster and used to host Docker images needed for its installation.
  address: ""                            # IP or hostname for your Docker registry. An internal registry will NOT be setup when this field is provided. Must be accessible from all the nodes in the cluster.
  port: 443                              # Port for your Docker registry.
  CA: ""                                 # Absolute path to the CA that was used when starting your Docker registry. The docker daemons on all nodes in the cluster will be configured with this CA.
add_ons:
  cni:
    disable: false
    provider: {{.CNI}}                     # Options: 'calico','weave','contiv','custom'. Selecting 'custom' will result in a CNI ready cluster, however it is up to you to configure a plugin after the install.
    options:
      calico:
        mode: overlay                    # Options: 'overlay','routed'. Routed pods can be addressed from outside the Kubernetes cluster; Overlay pods can only address each other.
  dns:
    disable: false
  heapster:
    disable: false
    options:
      heapster:
        replicas: 2
        service_type: ClusterIP          # Options: 'ClusterIP','NodePort','LoadBalancer','ExternalName'. Specify kubernetes ServiceType; default 'ClusterIP'
        sink: influxdb:http://heapster-influxdb.kube-system.svc:8086 # Specify the sink to store heapster data; default to a pod running on cluster.
      influxdb:
        pvc_name: ""                     # Provide the name of the persistent volume claim that you will create after installation. If not specified, the data will be stored in ephemeral storage.
  dashbard:
    disable: false
  package_manager:
    disable: false
    provider: helm                       # Options: 'helm'
etcd:
  expected_count: {{len .Etcd}}
  nodes:{{range .Etcd}}
  - host: {{.Host}}
    ip: {{.PublicIPv4}}
    internalip: {{.PrivateIPv4}}{{end}}
master:
  expected_count: {{len .Master}}
  nodes:{{range .Master}}
  - host: {{.Host}}
    ip: {{.PublicIPv4}}
    internalip: {{.PrivateIPv4}}{{end}}
  load_balanced_fqdn: {{.MasterNodeFQDN}}
  load_balanced_short_name: {{.MasterNodeShortName}}
worker:
  expected_count: {{len .Worker}}
  nodes:{{range .Worker}}
  - host: {{.Host}}
    ip: {{.PublicIPv4}}
    internalip: {{.PrivateIPv4}}{{end}}
ingress:
  expected_count: {{len .Ingress}}
  nodes:{{range .Ingress}}
  - host: {{.Host}}
    ip: {{.PublicIPv4}}
    internalip: {{.PrivateIPv4}}{{end}}
storage:
  expected_count: {{len .Storage}}
  nodes:{{range .Storage}}
  - host: {{.Host}}
    ip: {{.PublicIPv4}}
    internalip: {{.PrivateIPv4}}{{end}}
nfs:                                     # A set of NFS volumes for use by on-cluster persistent workloads, managed by Kismatic.
  nfs_volume: []
`
