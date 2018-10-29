package skytap

import (
	"context"
)

// Default URL paths
const (
	networksBasePath = "/v2/configurations/"
	networksPath     = "/networks"
)

// NetworksService is the contract for the services provided on the Skytap Network resource
type NetworksService interface {
	List(ctx context.Context, environmentID string) (*NetworkListResult, error)
	Get(ctx context.Context, environmentID string, id string) (*Network, error)
	Create(ctx context.Context, environmentID string, opts *CreateNetworkRequest) (*Network, error)
	Update(ctx context.Context, environmentID string, id string, network *UpdateNetworkRequest) (*Network, error)
	Delete(ctx context.Context, environmentID string, id string) error
}

// NetworksServiceClient is the NetworksService implementation
type NetworksServiceClient struct {
	client *Client
}

// Network is a network in the environment.
// Every environment can have multiple networks;
// the number of total networks that can be created is restricted by your customer account’s network quota.
type Network struct {
	ID                  *string         `json:"id"`
	URL                 *string         `json:"url"`
	Name                *string         `json:"name"`
	NetworkType         *string         `json:"network_type"`
	Subnet              *string         `json:"subnet"`
	SubnetAddr          *string         `json:"subnet_addr"`
	SubnetSize          *int            `json:"subnet_size"`
	Gateway             *string         `json:"gateway"`
	PrimaryNameserver   *string         `json:"primary_nameserver"`
	SecondaryNameserver *string         `json:"secondary_nameserver"`
	Region              *string         `json:"region"`
	Domain              *string         `json:"domain"`
	VPNAttachments      []VPNAttachment `json:"vpn_attachments"`
	Tunnelable          *bool           `json:"tunnelable"`
	Tunnels             []Tunnel        `json:"tunnels"`
}

// VPNAttachment are representations of the relationships between this network
// and any VPN or Private Network Connections it is attached to, including whether the network is currently connected.
type VPNAttachment struct {
	ID        *string               `json:"id"`
	Connected *bool                 `json:"connected"`
	Network   *VpnAttachmentNetwork `json:"network"`
	VPN       *VPN                  `json:"vpn"`
}

// VpnAttachmentNetwork describes the attachment network
type VpnAttachmentNetwork struct {
	ID              *string `json:"id"`
	Subnet          *string `json:"subnet"`
	NetworkName     *string `json:"network_name"`
	ConfigurationID *string `json:"configuration_id"`
}

// VPN describes a virtual private network.
type VPN struct {
	ID            *string `json:"id"`
	Name          *string `json:"name"`
	Enabled       *bool   `json:"enabled"`
	NatEnabled    *bool   `json:"nat_enabled"`
	RemoteSubnets *string `json:"remote_subnets"`
	RemotePeerIP  *string `json:"remote_peer_ip"`
	CanReconnect  *bool   `json:"can_reconnect"`
}

// Tunnel is a list of connections between this network and other networks
type Tunnel struct {
	ID            *string  `json:"id"`
	Status        *string  `json:"status"`
	Error         *string  `json:"error"`
	SourceNetwork *Network `json:"source_network"`
	TargetNetwork *Network `json:"target_network"`
}

// CreateNetworkRequest describes the create the network data
type CreateNetworkRequest struct {
	Name        *string      `json:"name"`
	NetworkType *NetworkType `json:"network_type"`
	Subnet      *string      `json:"subnet"`
	Domain      *string      `json:"domain"`
	Gateway     *string      `json:"gateway,omitempty"`
	Tunnelable  *bool        `json:"tunnelable,omitempty"`
}

// UpdateNetworkRequest describes the update the network data
type UpdateNetworkRequest struct {
	Name       *string `json:"name,omitempty"`
	Subnet     *string `json:"subnet,omitempty"`
	Domain     *string `json:"domain,omitempty"`
	Gateway    *string `json:"gateway,omitempty"`
	Tunnelable *bool   `json:"tunnelable,omitempty"`
}

// NetworkType is the type of network
type NetworkType string

// The architecture types
const (
	NetworkTypeAutomatic NetworkType = "automatic"
	NetworkTypeManual    NetworkType = "manual"
)

// NetworkListResult is the listing request specific struct
type NetworkListResult struct {
	Value []Network
}

// List the networks
func (s *NetworksServiceClient) List(ctx context.Context, environmentID string) (*NetworkListResult, error) {
	path := s.buildPath(environmentID, "")

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	err = s.client.setRequestListParameters(req, nil)
	if err != nil {
		return nil, err
	}

	var networkListResponse NetworkListResult
	_, err = s.client.do(ctx, req, &networkListResponse.Value)
	if err != nil {
		return nil, err
	}

	return &networkListResponse, nil
}

// Get a network
func (s *NetworksServiceClient) Get(ctx context.Context, environmentID string, id string) (*Network, error) {
	path := s.buildPath(environmentID, id)

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var network Network
	_, err = s.client.do(ctx, req, &network)
	if err != nil {
		return nil, err
	}

	return &network, nil
}

// Create a network
func (s *NetworksServiceClient) Create(ctx context.Context, environmentID string, opts *CreateNetworkRequest) (*Network, error) {
	path := s.buildPath(environmentID, "")

	req, err := s.client.newRequest(ctx, "POST", path, opts)
	if err != nil {
		return nil, err
	}

	var createdNetwork Network
	_, err = s.client.do(ctx, req, &createdNetwork)
	if err != nil {
		return nil, err
	}

	return &createdNetwork, nil
}

// Update a network
func (s *NetworksServiceClient) Update(ctx context.Context, environmentID string, id string, network *UpdateNetworkRequest) (*Network, error) {
	path := s.buildPath(environmentID, id)

	req, err := s.client.newRequest(ctx, "PUT", path, network)
	if err != nil {
		return nil, err
	}

	var updatedNetwork Network
	_, err = s.client.do(ctx, req, &updatedNetwork)
	if err != nil {
		return nil, err
	}

	return &updatedNetwork, nil
}

// Delete a network
func (s *NetworksServiceClient) Delete(ctx context.Context, environmentID string, id string) error {
	path := s.buildPath(environmentID, id)

	req, err := s.client.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *NetworksServiceClient) buildPath(environmentID string, networkID string) string {
	path := networksBasePath + environmentID + networksPath
	if networkID != "" {
		path += "/" + networkID
	}
	return path
}
