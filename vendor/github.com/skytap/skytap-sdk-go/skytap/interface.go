package skytap

import (
	"context"
	"fmt"
	"log"
)

// Default URL paths
const (
	interfacesBasePath = "/v2/configurations/"
	interfacesVMPath   = "/vms/"
	interfacesPath     = "/interfaces"
)

type interfacePathBuilder interface {
	Environment(string) interfacePathBuilder
	VM(string) interfacePathBuilder
	Interface(string) interfacePathBuilder
	Build() string
}

type interfacePathBuilderImpl struct {
	environment      string
	vm               string
	networkInterface string
}

func (pb *interfacePathBuilderImpl) Environment(environment string) interfacePathBuilder {
	pb.environment = environment
	return pb
}

func (pb *interfacePathBuilderImpl) VM(vm string) interfacePathBuilder {
	pb.vm = vm
	return pb
}

func (pb *interfacePathBuilderImpl) Interface(networkInterface string) interfacePathBuilder {
	pb.networkInterface = networkInterface
	return pb
}

func (pb *interfacePathBuilderImpl) Build() string {
	path := interfacesBasePath + pb.environment + interfacesVMPath + pb.vm + interfacesPath
	if pb.networkInterface != "" {
		return path + "/" + pb.networkInterface
	}
	return path
}

// InterfacesService is the contract for the services provided on the Skytap Interface resource
type InterfacesService interface {
	List(ctx context.Context, environmentID string, vmID string) (*InterfaceListResult, error)
	Get(ctx context.Context, environmentID string, vmID string, id string) (*Interface, error)
	Create(ctx context.Context, environmentID string, vmID string, nicType *CreateInterfaceRequest) (*Interface, error)
	Attach(ctx context.Context, environmentID string, vmID string, id string, networkID *AttachInterfaceRequest) (*Interface, error)
	Update(ctx context.Context, environmentID string, vmID string, id string, opt *UpdateInterfaceRequest) (*Interface, error)
	Delete(ctx context.Context, environmentID string, vmID string, id string) error
}

// InterfacesServiceClient is the InterfacesService implementation
type InterfacesServiceClient struct {
	client *Client
}

// Interface describes the VM's virtual network interface configuration
type Interface struct {
	ID                  *string              `json:"id"`
	IP                  *string              `json:"ip"`
	Hostname            *string              `json:"hostname"`
	MAC                 *string              `json:"mac"`
	ServicesCount       *int                 `json:"services_count"`
	Services            []PublishedService   `json:"services"`
	PublicIPsCount      *int                 `json:"public_ips_count"`
	PublicIPs           []map[string]string  `json:"public_ips"`
	VMID                *string              `json:"vm_id"`
	VMName              *string              `json:"vm_name"`
	Status              *string              `json:"status"`
	NetworkID           *string              `json:"network_id"`
	NetworkName         *string              `json:"network_name"`
	NetworkURL          *string              `json:"network_url"`
	NetworkType         *string              `json:"network_type"`
	NetworkSubnet       *string              `json:"network_subnet"`
	NICType             *NICType             `json:"nic_type"`
	SecondaryIPs        []SecondaryIP        `json:"secondary_ips"`
	PublicIPAttachments []PublicIPAttachment `json:"public_ip_attachments"`
}

// SecondaryIP holds secondary IP address data
type SecondaryIP struct {
	ID      *string `json:"id"`
	Address *string `json:"address"`
}

// PublicIPAttachment describes the public IP address data
type PublicIPAttachment struct {
	ID                    *int    `json:"id"`
	PublicIPAttachmentKey *int    `json:"public_ip_attachment_key"`
	Address               *string `json:"address"`
	ConnectType           *int    `json:"connect_type"`
	Hostname              *string `json:"hostname"`
	DNSName               *string `json:"dns_name"`
	PublicIPKey           *string `json:"public_ip_key"`
}

// NICType describes the different Network Interface Card types
type NICType string

// A list of the different NIC types
const (
	NICTypeDefault NICType = "default"
	NICTypePCNet32 NICType = "pcnet32"
	NICTypeE1000   NICType = "e1000"
	NICTypeE1000E  NICType = "e1000e"
	NICTypeVMXNet  NICType = "vmxnet"
	NICTypeVMXNet3 NICType = "vmxnet3"
)

// CreateInterfaceRequest describes the create the interface data
type CreateInterfaceRequest struct {
	NICType *NICType `json:"nic_type"`
}

// AttachInterfaceRequest configures the network id in order that the interface can be attached to a network
type AttachInterfaceRequest struct {
	NetworkID *string `json:"network_id"`
}

// UpdateInterfaceRequest describes the update the interface data
type UpdateInterfaceRequest struct {
	IP       *string `json:"ip,omitempty"`
	Hostname *string `json:"hostname,omitempty"`
}

// InterfaceListResult is the listing request specific struct
type InterfaceListResult struct {
	Value []Interface
}

// List the interfaces
func (s *InterfacesServiceClient) List(ctx context.Context, environmentID string, vmID string) (*InterfaceListResult, error) {
	var builder interfacePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Build()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	err = s.client.setRequestListParameters(req, nil)
	if err != nil {
		return nil, err
	}

	var interfaceListResponse InterfaceListResult
	_, err = s.client.do(ctx, req, &interfaceListResponse.Value, nil, nil)
	if err != nil {
		return nil, err
	}

	return &interfaceListResponse, nil
}

// Get an interface
func (s *InterfacesServiceClient) Get(ctx context.Context, environmentID string, vmID string, id string) (*Interface, error) {
	var builder interfacePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Interface(id).Build()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var networkInterface Interface
	_, err = s.client.do(ctx, req, &networkInterface, nil, nil)
	if err != nil {
		return nil, err
	}

	return &networkInterface, nil
}

// Create an interface
func (s *InterfacesServiceClient) Create(ctx context.Context, environmentID string, vmID string, nicType *CreateInterfaceRequest) (*Interface, error) {
	var builder interfacePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Build()

	req, err := s.client.newRequest(ctx, "POST", path, nicType)
	if err != nil {
		return nil, err
	}

	var createdInterface Interface
	_, err = s.client.do(ctx, req, &createdInterface, vmRequestRunStateStopped(environmentID, vmID), nicType)
	if err != nil {
		return nil, err
	}

	return &createdInterface, nil
}

// Attach an interface
func (s *InterfacesServiceClient) Attach(ctx context.Context, environmentID string, vmID string, id string, networkID *AttachInterfaceRequest) (*Interface, error) {
	var builder interfacePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Interface(id).Build()

	req, err := s.client.newRequest(ctx, "PUT", path, networkID)
	if err != nil {
		return nil, err
	}

	var updatedInterface Interface
	_, err = s.client.do(ctx, req, &updatedInterface, vmRunStateNotBusy(environmentID, vmID), networkID)
	if err != nil {
		return nil, err
	}

	return &updatedInterface, nil
}

// Update an interface
func (s *InterfacesServiceClient) Update(ctx context.Context, environmentID string, vmID string, id string, opts *UpdateInterfaceRequest) (*Interface, error) {
	var builder interfacePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Interface(id).Build()

	req, err := s.client.newRequest(ctx, "PUT", path, opts)
	if err != nil {
		return nil, err
	}

	var updatedInterface Interface
	_, err = s.client.do(ctx, req, &updatedInterface, vmRequestRunStateStopped(environmentID, vmID), opts)
	if err != nil {
		return nil, err
	}

	return &updatedInterface, nil
}

// Delete an interface
func (s *InterfacesServiceClient) Delete(ctx context.Context, environmentID string, vmID string, id string) error {
	var builder interfacePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Interface(id).Build()

	req, err := s.client.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil, vmRequestRunStateStopped(environmentID, vmID), nil)
	if err != nil {
		return err
	}

	return nil
}

func (payload *CreateInterfaceRequest) compareResponse(ctx context.Context, c *Client, v interface{}, state *environmentVMRunState) (string, bool) {
	if interfaceOriginal, ok := v.(*Interface); ok {
		adapter, err := c.Interfaces.Get(ctx, *state.environmentID, *state.vmID, *interfaceOriginal.ID)
		if err != nil {
			return requestNotAsExpected, false
		}
		actual := CreateInterfaceRequest{
			adapter.NICType,
		}
		if payload.string() == actual.string() {
			return "", true
		}
		return "network adapter not ready", false
	}
	log.Printf("[ERROR] SDK interface comparison not possible on (%v)\n", v)
	return requestNotAsExpected, false
}

func (payload *AttachInterfaceRequest) compareResponse(ctx context.Context, c *Client, v interface{}, state *environmentVMRunState) (string, bool) {
	if interfaceOriginal, ok := v.(*Interface); ok {
		adapter, err := c.Interfaces.Get(ctx, *state.environmentID, *state.vmID, *interfaceOriginal.ID)
		if err != nil {
			return requestNotAsExpected, false
		}
		actual := AttachInterfaceRequest{
			adapter.NetworkID,
		}
		if payload.string() == actual.string() {
			return "", true
		}
		return "network adapter not ready", false
	}
	log.Printf("[ERROR] SDK interface comparison not possible on (%v)\n", v)
	return requestNotAsExpected, false
}

func (payload *UpdateInterfaceRequest) compareResponse(ctx context.Context, c *Client, v interface{}, state *environmentVMRunState) (string, bool) {
	if interfaceOriginal, ok := v.(*Interface); ok {
		adapter, err := c.Interfaces.Get(ctx, *state.environmentID, *state.vmID, *interfaceOriginal.ID)
		if err != nil {
			return requestNotAsExpected, false
		}
		actual := UpdateInterfaceRequest{
			adapter.IP,
			adapter.Hostname,
		}
		if payload.string() == actual.string() {
			return "", true
		}
		return "network adapter not ready", false
	}
	log.Printf("[ERROR] SDK interface comparison not possible on (%v)\n", v)
	return requestNotAsExpected, false
}

func (payload *CreateInterfaceRequest) string() string {
	nicType := ""

	if payload.NICType != nil {
		nicType = string(*payload.NICType)
	}
	s := fmt.Sprintf("%s",
		nicType)
	log.Printf("[DEBUG] SDK create interface payload (%s)\n", s)
	return s
}

func (payload *AttachInterfaceRequest) string() string {
	networkID := ""

	if payload.NetworkID != nil {
		networkID = *payload.NetworkID
	}
	s := fmt.Sprintf("%s",
		networkID)
	log.Printf("[DEBUG] SDK attach interface payload (%s)\n", s)
	return s
}

func (payload *UpdateInterfaceRequest) string() string {
	ip := ""
	hostname := ""

	if payload.IP != nil {
		ip = string(*payload.IP)
	}
	if payload.Hostname != nil {
		hostname = string(*payload.Hostname)
	}
	s := fmt.Sprintf("%s%s",
		ip,
		hostname)
	log.Printf("[DEBUG] SDK update interface payload (%s)\n", s)
	return s
}
