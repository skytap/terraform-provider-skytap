package skytap

import (
	"context"
	"fmt"
	"log"
)

// Default URL paths
const (
	publishedServicesBasePath       = "/v2/configurations/"
	publishedServicesVMPath         = "/vms/"
	publishedServicesInterfacesPath = "/interfaces/"
	publishedServicesPath           = "/services"
)

type publishedServicePathBuilder interface {
	Environment(string) publishedServicePathBuilder
	VM(string) publishedServicePathBuilder
	Interface(string) publishedServicePathBuilder
	Service(string) publishedServicePathBuilder
	Build() string
}

type publishedServicePathBuilderImpl struct {
	environment      string
	vm               string
	networkInterface string
	publishedService string
}

func (pb *publishedServicePathBuilderImpl) Environment(environment string) publishedServicePathBuilder {
	pb.environment = environment
	return pb
}

func (pb *publishedServicePathBuilderImpl) VM(vm string) publishedServicePathBuilder {
	pb.vm = vm
	return pb
}

func (pb *publishedServicePathBuilderImpl) Interface(networkInterface string) publishedServicePathBuilder {
	pb.networkInterface = networkInterface
	return pb
}

func (pb *publishedServicePathBuilderImpl) Service(publishedService string) publishedServicePathBuilder {
	pb.publishedService = publishedService
	return pb
}

func (pb *publishedServicePathBuilderImpl) Build() string {
	path := publishedServicesBasePath + pb.environment + publishedServicesVMPath + pb.vm + publishedServicesInterfacesPath + pb.networkInterface + publishedServicesPath
	if pb.publishedService != "" {
		return path + "/" + pb.publishedService
	}
	return path
}

// PublishedServicesService is the contract for the services provided on the Skytap PublishedServices resource
type PublishedServicesService interface {
	List(ctx context.Context, environmentID string, vmID string, nicID string) (*PublishedServiceListResult, error)
	Get(ctx context.Context, environmentID string, vmID string, nicID string, id string) (*PublishedService, error)
	Create(ctx context.Context, environmentID string, vmID string, nicID string, internalPort *CreatePublishedServiceRequest) (*PublishedService, error)
	Update(ctx context.Context, environmentID string, vmID string, nicID string, id string, internalPort *UpdatePublishedServiceRequest) (*PublishedService, error)
	Delete(ctx context.Context, environmentID string, vmID string, nicID string, id string) error
}

// PublishedServicesServiceClient is the PublishedServicesService implementation
type PublishedServicesServiceClient struct {
	client *Client
}

// PublishedService describes a publishedService provided on the connected network
type PublishedService struct {
	ID           *string `json:"id"`
	InternalPort *int    `json:"internal_port"`
	ExternalIP   *string `json:"external_ip"`
	ExternalPort *int    `json:"external_port"`
	Name         *string `json:"name,omitempty"`
}

// CreatePublishedServiceRequest describes the create the publishedService data
type CreatePublishedServiceRequest struct {
	InternalPort *int `json:"internal_port"`
}

// UpdatePublishedServiceRequest describes the update the publishedService data
type UpdatePublishedServiceRequest struct {
	CreatePublishedServiceRequest
}

// PublishedServiceListResult is the listing request specific struct
type PublishedServiceListResult struct {
	Value []PublishedService
}

// List the services
func (s *PublishedServicesServiceClient) List(ctx context.Context, environmentID string, vmID string, nicID string) (*PublishedServiceListResult, error) {
	var builder publishedServicePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Interface(nicID).Build()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	err = s.client.setRequestListParameters(req, nil)
	if err != nil {
		return nil, err
	}

	var serviceListResponse PublishedServiceListResult
	_, err = s.client.do(ctx, req, &serviceListResponse.Value, nil, nil)
	if err != nil {
		return nil, err
	}

	return &serviceListResponse, nil
}

// Get a publishedService
func (s *PublishedServicesServiceClient) Get(ctx context.Context, environmentID string, vmID string, nicID string, id string) (*PublishedService, error) {
	var builder publishedServicePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Interface(nicID).Service(id).Build()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var service PublishedService
	_, err = s.client.do(ctx, req, &service, nil, nil)
	if err != nil {
		return nil, err
	}

	return &service, nil
}

// Create a publishedService
func (s *PublishedServicesServiceClient) Create(ctx context.Context, environmentID string, vmID string, nicID string, internalPort *CreatePublishedServiceRequest) (*PublishedService, error) {
	var builder publishedServicePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Interface(nicID).Build()

	req, err := s.client.newRequest(ctx, "POST", path, internalPort)
	if err != nil {
		return nil, err
	}
	state := vmRunStateNotBusy(environmentID, vmID)
	state.adapterID = strToPtr(nicID)
	var createdService PublishedService
	_, err = s.client.do(ctx, req, &createdService, state, internalPort)
	if err != nil {
		return nil, err
	}

	return &createdService, nil
}

// Update a publishedService
func (s *PublishedServicesServiceClient) Update(ctx context.Context, environmentID string, vmID string, nicID string, id string, internalPort *UpdatePublishedServiceRequest) (*PublishedService, error) {
	err := s.Delete(ctx, environmentID, vmID, nicID, id)
	if err != nil {
		return nil, err
	}
	return s.Create(ctx, environmentID, vmID, nicID, &internalPort.CreatePublishedServiceRequest)
}

// Delete a publishedService
func (s *PublishedServicesServiceClient) Delete(ctx context.Context, environmentID string, vmID string, nicID string, id string) error {
	var builder publishedServicePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Interface(nicID).Service(id).Build()

	req, err := s.client.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (payload *CreatePublishedServiceRequest) compare(ctx context.Context, c *Client, v interface{}, state *environmentVMRunState) (string, bool) {
	if serviceOriginal, ok := v.(*PublishedService); ok {
		publishedService, err := c.PublishedServices.Get(ctx, *state.environmentID, *state.vmID, *state.adapterID, *serviceOriginal.ID)
		if err != nil {
			return requestNotAsExpected, false
		}
		actual := CreatePublishedServiceRequest{
			publishedService.InternalPort,
		}
		if payload.string() == actual.string() {
			return "", true
		}
		return "published service not ready", false
	}
	log.Printf("[ERROR] SDK published service comparison not possible on (%v)\n", v)
	return requestNotAsExpected, false
}

func (payload *UpdatePublishedServiceRequest) compare(ctx context.Context, c *Client, v interface{}, state *environmentVMRunState) (string, bool) {
	return payload.CreatePublishedServiceRequest.compare(ctx, c, v, state)
}

func (payload *CreatePublishedServiceRequest) string() string {
	internalPort := ""

	if payload.InternalPort != nil {
		internalPort = fmt.Sprintf("%d", *payload.InternalPort)
	}
	s := fmt.Sprintf("%s",
		internalPort)
	log.Printf("[DEBUG] SDK create published service payload (%s)\n", s)
	return s
}

func (payload *UpdatePublishedServiceRequest) string() string {
	return payload.string()
}
