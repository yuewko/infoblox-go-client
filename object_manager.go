package ibclient

import (
	"fmt"
	"regexp"
)

type ObjectManager struct {
	connector *Connector
	dockerID  string
}

func NewObjectManager(connector *Connector, dockerID string) *ObjectManager {
	objMgr := new(ObjectManager)

	objMgr.connector = connector
	objMgr.dockerID = dockerID

	return objMgr
}

func (objMgr *ObjectManager) getBasicEA(cloudApiOwned Bool) EA {
	ea := make(EA)
	ea["Cloud API Owned"] = cloudApiOwned
	ea["CMP Type"] = "Docker"
	ea["Tenant ID"] = objMgr.dockerID
	return ea
}

func (objMgr *ObjectManager) CreateNetworkView(name string) (*NetworkView, error) {
	networkView := NewNetworkView(NetworkView{
		Name: name,
		Ea:   objMgr.getBasicEA(false)})

	ref, err := objMgr.connector.CreateObject(networkView)
	networkView.Ref = ref

	return networkView, err
}

func (objMgr *ObjectManager) CreateDefaultNetviews(globalNetview string, localNetview string) (globalNetviewRef string, localNetviewRef string, err error) {
	globalNetviewRef = ""
	localNetviewRef = ""

	var globalNetviewObj *NetworkView
	if globalNetviewObj, err = objMgr.GetNetworkView(globalNetview); err != nil {
		return
	}
	if globalNetviewObj == nil {
		if globalNetviewObj, err = objMgr.CreateNetworkView(globalNetview); err != nil {
			return
		}
	}
	globalNetviewRef = globalNetviewObj.Ref

	var localNetviewObj *NetworkView
	if localNetviewObj, err = objMgr.GetNetworkView(localNetview); err != nil {
		return
	}
	if localNetviewObj == nil {
		if localNetviewObj, err = objMgr.CreateNetworkView(localNetview); err != nil {
			return
		}
	}
	localNetviewRef = localNetviewObj.Ref

	return
}

func (objMgr *ObjectManager) CreateNetwork(netview string, cidr string) (*Network, error) {
	network := NewNetwork(Network{
		NetviewName: netview,
		Cidr:        cidr,
		Ea:          objMgr.getBasicEA(true)})

	ref, err := objMgr.connector.CreateObject(network)
	network.Ref = ref

	return network, err
}

func (objMgr *ObjectManager) CreateNetworkContainer(netview string, cidr string) (*NetworkContainer, error) {
	container := NewNetworkContainer(NetworkContainer{
		NetviewName: netview,
		Cidr:        cidr,
		Ea:          objMgr.getBasicEA(true)})

	ref, err := objMgr.connector.CreateObject(container)
	container.Ref = ref

	return container, err
}

func (objMgr *ObjectManager) GetNetworkView(name string) (*NetworkView, error) {
	var res []NetworkView

	netview := NewNetworkView(NetworkView{Name: name})

	err := objMgr.connector.GetObject(netview, "", &res)

	if err != nil || res == nil || len(res) == 0 {
		return nil, err
	}

	return &res[0], nil
}

func BuildNetworkViewFromRef(ref string) *NetworkView {
	// networkview/ZG5zLm5ldHdvcmtfdmlldyQyMw:global_view/false
	r := regexp.MustCompile(`networkview/\w+:([^/]+)/\w+`)
	m := r.FindStringSubmatch(ref)

	if m == nil {
		return nil
	}

	return &NetworkView{
		Ref:  ref,
		Name: m[1],
	}
}

func BuildNetworkFromRef(ref string) *Network {
	// network/ZG5zLm5ldHdvcmskODkuMC4wLjAvMjQvMjU:89.0.0.0/24/global_view
	r := regexp.MustCompile(`network/\w+:(\d+\.\d+\.\d+\.\d+/\d+)/(.+)`)
	m := r.FindStringSubmatch(ref)

	if m == nil {
		return nil
	}

	return &Network{
		Ref:         ref,
		NetviewName: m[2],
		Cidr:        m[1],
	}
}

func (objMgr *ObjectManager) GetNetwork(netview string, cidr string) (*Network, error) {
	var res []Network

	network := NewNetwork(Network{
		NetviewName: netview,
		Cidr:        cidr})

	err := objMgr.connector.GetObject(network, "", &res)

	if err != nil || res == nil || len(res) == 0 {
		return nil, err
	}

	return &res[0], nil
}

func (objMgr *ObjectManager) GetNetworkContainer(netview string, cidr string) (*NetworkContainer, error) {
	var res []NetworkContainer

	nwcontainer := NewNetworkContainer(NetworkContainer{
		NetviewName: netview,
		Cidr:        cidr})

	err := objMgr.connector.GetObject(nwcontainer, "", &res)

	if err != nil || res == nil || len(res) == 0 {
		return nil, err
	}

	return &res[0], nil
}

func GetIPAddressFromRef(ref string) string {
	// fixedaddress/ZG5zLmJpbmRfY25h:12.0.10.1/external
	r := regexp.MustCompile(`fixedaddress/\w+:(\d+\.\d+\.\d+\.\d+)/.+`)
	m := r.FindStringSubmatch(ref)

	if m != nil {
		return m[1]
	}
	return ""
}

func (objMgr *ObjectManager) AllocateIP(netview string, cidr string, macAddress string, vmID string) (*FixedAddress, error) {
	if len(macAddress) == 0 {
		macAddress = "00:00:00:00:00:00"
	}

	ea := objMgr.getBasicEA(true)
	if vmID == "" {
		vmID = "N/A"
	}
	ea["VM ID"] = vmID

	fixedAddr := NewFixedAddress(FixedAddress{
		NetviewName: netview,
		Cidr:        cidr,
		IPAddress:   fmt.Sprintf("func:nextavailableip:%s,%s", cidr, netview),
		Mac:         macAddress,
		Ea:          ea})

	ref, err := objMgr.connector.CreateObject(fixedAddr)
	fixedAddr.Ref = ref
	fixedAddr.IPAddress = GetIPAddressFromRef(ref)

	return fixedAddr, err
}

func (objMgr *ObjectManager) AllocateNetwork(netview string, cidr string, prefixLen uint) (network *Network, err error) {
	network = nil

	networkReq := NewNetwork(Network{
		NetviewName: netview,
		Cidr:        fmt.Sprintf("func:nextavailablenetwork:%s,%s,%d", cidr, netview, prefixLen),
		Ea:          objMgr.getBasicEA(true)})

	ref, err := objMgr.connector.CreateObject(networkReq)
	if err == nil && len(ref) > 0 {
		network = BuildNetworkFromRef(ref)
	}

	return
}

func (objMgr *ObjectManager) GetFixedAddress(netview string, ipAddr string) (*FixedAddress, error) {
	var res []FixedAddress

	fixedAddr := NewFixedAddress(FixedAddress{
		NetviewName: netview,
		IPAddress:   ipAddr})

	err := objMgr.connector.GetObject(fixedAddr, "", &res)

	if err != nil || res == nil || len(res) == 0 {
		return nil, err
	}

	return &res[0], nil
}

func (objMgr *ObjectManager) ReleaseIP(netview string, ipAddr string) (string, error) {
	fixAddress, _ := objMgr.GetFixedAddress(netview, ipAddr)

	return objMgr.connector.DeleteObject(fixAddress.Ref)
}

func (objMgr *ObjectManager) DeleteLocalNetwork(ref string, localNetview string) (string, error) {
	network := BuildNetworkFromRef(ref)
	if network != nil && network.NetviewName == localNetview {
		return objMgr.connector.DeleteObject(ref)
	}

	return "", nil
}

func (objMgr *ObjectManager) GetEADefinition(name string) (*EADefinition, error) {
	var res []EADefinition

	eadef := NewEADefinition(EADefinition{Name: name})

	err := objMgr.connector.GetObject(eadef, "", &res)

	if err != nil || res == nil || len(res) == 0 {
		return nil, err
	}

	return &res[0], nil
}

func (objMgr *ObjectManager) CreateEADefinition(eadef EADefinition) (*EADefinition, error) {
	newEadef := NewEADefinition(eadef)

	ref, err := objMgr.connector.CreateObject(newEadef)
	newEadef.Ref = ref

	return newEadef, err
}
