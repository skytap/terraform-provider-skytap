package skytap

import (
	"context"
	"fmt"
)

// Default URL paths
const (
	environmentLegacyBasePath = "/configurations"
	environmentBasePath       = "/v2/configurations"
)

// EnvironmentsService is the contract for the services provided on the Skytap Environment resource
type EnvironmentsService interface {
	List(ctx context.Context) (*EnvironmentListResult, error)
	Get(ctx context.Context, id string) (*Environment, error)
	Create(ctx context.Context, createEnvironmentRequest *CreateEnvironmentRequest) (*Environment, error)
	Update(ctx context.Context, id string, updateEnvironmentRequest *UpdateEnvironmentRequest) (*Environment, error)
	Delete(ctx context.Context, id string) error
}

// EnvironmentsServiceClient is the EnvironmentsService implementation
type EnvironmentsServiceClient struct {
	client *Client
}

// Environment resource struct definitions.
type Environment struct {
	ID                      *string              `json:"id"`
	URL                     *string              `json:"url"`
	Name                    *string              `json:"name"`
	Description             *string              `json:"description"`
	Errors                  []string             `json:"errors"`
	ErrorDetails            []string             `json:"error_details"`
	Runstate                *EnvironmentRunstate `json:"runstate"`
	RateLimited             *bool                `json:"rate_limited"`
	LastRun                 *string              `json:"last_run"`
	SuspendOnIdle           *int                 `json:"suspend_on_idle"`
	SuspendAtTime           *string              `json:"suspend_at_time"`
	OwnerURL                *string              `json:"owner_url"`
	OwnerName               *string              `json:"owner_name"`
	OwnerID                 *string              `json:"owner_id"`
	VMCount                 *int                 `json:"vm_count"`
	Storage                 *int                 `json:"storage"`
	NetworkCount            *int                 `json:"network_count"`
	CreatedAt               *string              `json:"created_at"`
	Region                  *string              `json:"region"`
	RegionBackend           *string              `json:"region_backend"`
	SVMs                    *int                 `json:"svms"`
	CanSaveAsTemplate       *bool                `json:"can_save_as_template"`
	CanCopy                 *bool                `json:"can_copy"`
	CanDelete               *bool                `json:"can_delete"`
	CanChangeState          *bool                `json:"can_change_state"`
	CanShare                *bool                `json:"can_share"`
	CanEdit                 *bool                `json:"can_edit"`
	LabelCount              *int                 `json:"label_count"`
	LabelCategoryCount      *int                 `json:"label_category_count"`
	CanTag                  *bool                `json:"can_tag"`
	Tags                    []Tag                `json:"tags"`
	TagList                 *string              `json:"tag_list"`
	Alerts                  []Alert              `json:"alerts"`
	PublishedServiceCount   *int                 `json:"published_service_count"`
	PublicIPCount           *int                 `json:"public_ip_count"`
	AutoSuspendDescription  *string              `json:"auto_suspend_description"`
	Stages                  []Stage              `json:"stages"`
	StagedExecution         *StagedExecution     `json:"staged_execution"`
	SequencingEnabled       *bool                `json:"sequencing_enabled"`
	NoteCount               *int                 `json:"note_count"`
	ProjectCountForUser     *int                 `json:"project_count_for_user"`
	ProjectCount            *int                 `json:"project_count"`
	PublishSetCount         *int                 `json:"publish_set_count"`
	ScheduleCount           *int                 `json:"schedule_count"`
	VpnCount                *int                 `json:"vpn_count"`
	OutboundTraffic         *bool                `json:"outbound_traffic"`
	Routable                *bool                `json:"routable"`
	VMs                     []VM                 `json:"vms"`
	Networks                []Network            `json:"networks"`
	ContainersCount         *int                 `json:"containers_count"`
	ContainerHostsCount     *int                 `json:"container_hosts_count"`
	PlatformErrors          []string             `json:"platform_errors"`
	SVMsByArchitecture      *SVMsByArchitecture  `json:"svms_by_architecture"`
	AllVmsSupportSuspend    *bool                `json:"all_vms_support_suspend"`
	ShutdownOnIdle          *int                 `json:"shutdown_on_idle"`
	ShutdownAtTime          *string              `json:"shutdown_at_time"`
	AutoShutdownDescription *string              `json:"auto_shutdown_description"`
}

// Tag describes environment tag data
type Tag struct {
	ID    *string `json:"id"`
	Value *string `json:"value"`
}

// Alert describes an environment alert
type Alert struct {
	ID                   string `json:"id"`
	DisplayType          string `json:"display_type"`
	Dismissable          bool   `json:"dismissable"`
	Message              string `json:"message"`
	DisplayOnGeneral     bool   `json:"display_on_general"`
	DisplayOnLogin       bool   `json:"display_on_login"`
	DisplayOnSmartclient bool   `json:"display_on_smartclient"`
}

// Stage describes the VM stage sequence
type Stage struct {
	DelayAfterFinishSeconds *int     `json:"delay_after_finish_seconds"`
	Index                   *int     `json:"index"`
	VMIDs                   []string `json:"vm_ids"`
}

// StagedExecution describes the status of a running VM sequence
type StagedExecution struct {
	ActionType                          *string  `json:"action_type"`
	CurrentStageDelayAfterFinishSeconds *int     `json:"current_stage_delay_after_finish_seconds"`
	CurrentStageIndex                   *int     `json:"current_stage_index"`
	CurrentStageFinishedAt              *string  `json:"current_stage_finished_at"`
	VMIDs                               []string `json:"vm_ids"`
}

// SVMsByArchitecture lists the number of x86 and power SVMs consumed by VMs in the environment
type SVMsByArchitecture struct {
	X86   *int `json:"x86"`
	Power *int `json:"power"`
}

// EnvironmentRunstate enumerates the possible environment running states
type EnvironmentRunstate string

// The environment running states
const (
	EnvironmentRunstateStopped   EnvironmentRunstate = "stopped"
	EnvironmentRunstateSuspended EnvironmentRunstate = "suspended"
	EnvironmentRunstateRunning   EnvironmentRunstate = "running"
	EnvironmentRunstateBusy      EnvironmentRunstate = "busy"
)

// EnvironmentListResult is the list request specific struct
type EnvironmentListResult struct {
	Value []Environment
}

// CreateEnvironmentRequest describes the create the environment data
type CreateEnvironmentRequest struct {
	TemplateID      *string `json:"template_id,omitempty"`
	ProjectID       *int    `json:"project_id,omitempty"`
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	Owner           *string `json:"owner,omitempty"`
	OutboundTraffic *bool   `json:"outbound_traffic,omitempty"`
	Routable        *bool   `json:"routable,omitempty"`
	SuspendOnIdle   *int    `json:"suspend_on_idle,omitempty"`
	SuspendAtTime   *string `json:"suspend_at_time,omitempty"`
	ShutdownOnIdle  *int    `json:"shutdown_on_idle,omitempty"`
	ShutdownAtTime  *string `json:"shutdown_at_time,omitempty"`
}

// UpdateEnvironmentRequest describes the update the environment data
type UpdateEnvironmentRequest struct {
	Name            *string              `json:"name,omitempty"`
	Description     *string              `json:"description,omitempty"`
	Owner           *string              `json:"owner,omitempty"`
	OutboundTraffic *bool                `json:"outbound_traffic,omitempty"`
	Routable        *bool                `json:"routable,omitempty"`
	SuspendOnIdle   *int                 `json:"suspend_on_idle,omitempty"`
	SuspendAtTime   *string              `json:"suspend_at_time,omitempty"`
	ShutdownOnIdle  *int                 `json:"shutdown_on_idle,omitempty"`
	ShutdownAtTime  *string              `json:"shutdown_at_time,omitempty"`
	Runstate        *EnvironmentRunstate `json:"runstate,omitempty"`
}

// List the environments
func (s *EnvironmentsServiceClient) List(ctx context.Context) (*EnvironmentListResult, error) {
	req, err := s.client.newRequest(ctx, "GET", environmentBasePath, nil)
	if err != nil {
		return nil, err
	}

	err = s.client.setRequestListParameters(req, nil)
	if err != nil {
		return nil, err
	}

	var environmentsListResponse EnvironmentListResult
	_, err = s.client.do(ctx, req, &environmentsListResponse.Value)
	if err != nil {
		return nil, err
	}

	return &environmentsListResponse, nil
}

// Get an environment
func (s *EnvironmentsServiceClient) Get(ctx context.Context, id string) (*Environment, error) {
	path := fmt.Sprintf("%s/%s", environmentBasePath, id)

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var environment Environment
	_, err = s.client.do(ctx, req, &environment)
	if err != nil {
		return nil, err
	}

	return &environment, nil
}

// Create an environment
func (s *EnvironmentsServiceClient) Create(ctx context.Context, request *CreateEnvironmentRequest) (*Environment, error) {
	req, err := s.client.newRequest(ctx, "POST", environmentLegacyBasePath, request)
	if err != nil {
		return nil, err
	}

	var createdEnvironment Environment
	_, err = s.client.do(ctx, req, &createdEnvironment)
	if err != nil {
		return nil, err
	}

	runstate := EnvironmentRunstateRunning

	updateOpts := &UpdateEnvironmentRequest{
		Name:            request.Name,
		Description:     request.Description,
		Owner:           request.Owner,
		OutboundTraffic: request.OutboundTraffic,
		Routable:        request.Routable,
		SuspendOnIdle:   request.SuspendOnIdle,
		SuspendAtTime:   request.SuspendAtTime,
		ShutdownOnIdle:  request.ShutdownOnIdle,
		ShutdownAtTime:  request.ShutdownAtTime,
		Runstate:        &runstate,
	}

	// update environment after creation to establish the resource information.
	environment, err := s.Update(ctx, ptrToStr(createdEnvironment.ID), updateOpts)
	if err != nil {
		return nil, err
	}

	return environment, nil
}

// Update an environment
func (s *EnvironmentsServiceClient) Update(ctx context.Context, id string, updateEnvironment *UpdateEnvironmentRequest) (*Environment, error) {
	path := fmt.Sprintf("%s/%s", environmentBasePath, id)

	req, err := s.client.newRequest(ctx, "PUT", path, updateEnvironment)
	if err != nil {
		return nil, err
	}

	var environment Environment
	_, err = s.client.do(ctx, req, &environment)
	if err != nil {
		return nil, err
	}

	return &environment, nil
}

// Delete an environment
func (s *EnvironmentsServiceClient) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", environmentLegacyBasePath, id)

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
