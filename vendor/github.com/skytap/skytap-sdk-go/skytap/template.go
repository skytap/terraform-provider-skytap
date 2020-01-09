package skytap

import (
	"context"
	"fmt"
)

// Default URL paths
const (
	templateBasePath = "/v2/templates"
)

// TemplatesService is the contract for the services provided on the Skytap Template resource
type TemplatesService interface {
	List(ctx context.Context) (*TemplateListResult, error)
	Get(ctx context.Context, id string) (*Template, error)
}

// TemplatesServiceClient is the TemplatesService implementation
type TemplatesServiceClient struct {
	client *Client
}

// Template resource struct definitions.
type Template struct {
	ID                  *string             `json:"id"`
	URL                 *string             `json:"url"`
	Name                *string             `json:"name"`
	Errors              []string            `json:"errors"`
	Busy                *bool               `json:"busy"`
	Public              *bool               `json:"public"`
	Description         *string             `json:"description"`
	VMCount             *int                `json:"vm_count"`
	Storage             *int                `json:"storage"`
	NetworkCount        *int                `json:"network_count"`
	CreatedAt           *string             `json:"created_at"`
	Region              *string             `json:"region"`
	RegionBackend       *string             `json:"region_backend"`
	SVMs                *int                `json:"svms"`
	LastInstalled       *string             `json:"last_installed"`
	CanCopy             *bool               `json:"can_copy"`
	CanDelete           *bool               `json:"can_delete"`
	CanShare            *bool               `json:"can_share"`
	LabelCount          *int                `json:"label_count"`
	LabelCategoryCount  *int                `json:"label_category_count"`
	CanTag              *bool               `json:"can_tag"`
	Tags                []Tag               `json:"tags"`
	TagList             *string             `json:"tag_list"`
	ProjectCountForUser *int                `json:"project_count_for_user"`
	ProjectCount        *int                `json:"project_count"`
	ContainersCount     *int                `json:"containers_count"`
	ContainerHostsCount *int                `json:"container_hosts_count"`
	VMs                 []VM                `json:"vms"`
	Networks            []Network           `json:"networks"`
	SVMsByArchitecture  *SVMsByArchitecture `json:"svms_by_architecture"`
}

// TemplateListResult is the list request specific struct
type TemplateListResult struct {
	Value []Template
}

// List the templates
func (s *TemplatesServiceClient) List(ctx context.Context) (*TemplateListResult, error) {
	req, err := s.client.newRequest(ctx, "GET", templateBasePath, nil)
	if err != nil {
		return nil, err
	}

	err = s.client.setRequestListParameters(req, nil)
	if err != nil {
		return nil, err
	}

	var templatesListResponse TemplateListResult
	_, err = s.client.do(ctx, req, &templatesListResponse.Value, nil, nil)
	if err != nil {
		return nil, err
	}

	return &templatesListResponse, nil
}

// Get an template
func (s *TemplatesServiceClient) Get(ctx context.Context, id string) (*Template, error) {
	path := fmt.Sprintf("%s/%s", templateBasePath, id)

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var template Template
	_, err = s.client.do(ctx, req, &template, nil, nil)
	if err != nil {
		return nil, err
	}

	return &template, nil
}
