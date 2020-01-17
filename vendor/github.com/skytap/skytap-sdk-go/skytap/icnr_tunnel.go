package skytap

import (
	"context"
	"fmt"
)

// Default URL paths
const (
	tunnelBasePath = "/v2/tunnels"
)

type ICNRTunnelService interface {
	Get(ctx context.Context, id string) (*ICNRTunnel, error)
	Create(ctx context.Context, source int, target int) (*ICNRTunnel, error)
	Delete(ctx context.Context, id string) error
}

// INCRTunnel model a Inter-Configuration network router
type ICNRTunnel struct {
	ID     *string  `json:"id"`
	Status *string  `json:"status"`
	Error  *string  `json:"error"`
	Source *Network `json:"source_network"`
	Target *Network `json:"target_network"`
}

type icnrTunnelCreate struct {
	Source int `json:"source_network_id"`
	Target int `json:"target_network_id"`
}

// ICNRTunnelClient is the ICNRTunnelService implementation
type ICNRTunnelClient struct {
	client *Client
}

// Get a ICNR Tunnel
func (s *ICNRTunnelClient) Get(ctx context.Context, id string) (*ICNRTunnel, error) {
	path := fmt.Sprintf("%s/%s.json", tunnelBasePath, id)
	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var tunnel ICNRTunnel
	_, err = s.client.do(ctx, req, &tunnel, nil, nil)
	if err != nil {
		return nil, err
	}

	return &tunnel, nil
}

// Create a ICNR tunnel
func (s *ICNRTunnelClient) Create(ctx context.Context, source int, target int) (*ICNRTunnel, error) {
	createBody := icnrTunnelCreate{
		Source: source,
		Target: target,
	}
	req, err := s.client.newRequest(ctx, "POST", tunnelBasePath, createBody)
	if err != nil {
		return nil, err
	}

	var tunnel ICNRTunnel
	_, err = s.client.do(ctx, req, &tunnel, nil, nil)
	if err != nil {
		return nil, err
	}
	return &tunnel, nil

}

// Delete a ICNR tunnel
func (s *ICNRTunnelClient) Delete(ctx context.Context, id string) (err error) {
	path := fmt.Sprintf("%s/%s.json", tunnelBasePath, id)
	req, err := s.client.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return
	}

	_, err = s.client.do(ctx, req, nil, nil, nil)
	return
}
