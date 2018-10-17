package skytap

import (
	"context"
	"fmt"
)

// Default URL paths
const (
	projectsLegacyBasePath = "/projects"
	projectsBasePath       = "/v2/projects"
)

// ProjectsService is the contract for the services provided on the Skytap Project resource
type ProjectsService interface {
	List(ctx context.Context) (*ProjectListResult, error)
	Get(ctx context.Context, id int) (*Project, error)
	Create(ctx context.Context, project *Project) (*Project, error)
	Update(ctx context.Context, id int, project *Project) (*Project, error)
	Delete(ctx context.Context, id int) error
}

// ProjectsServiceClient is the ProjectsService implementation
type ProjectsServiceClient struct {
	client *Client
}

// Project resource struct definitions
type Project struct {
	ID                 *int         `json:"id,omitempty,string"`
	Name               *string      `json:"name,omitempty"`
	Summary            *string      `json:"summary,omitempty"`
	AutoAddRoleName    *ProjectRole `json:"auto_add_role_name,omitempty"`
	ShowProjectMembers *bool        `json:"show_project_members,omitempty"`
}

// ProjectRole is the enumeration of the different possible project roles
type ProjectRole string

// The different project roles
const (
	ProjectRoleViewer      ProjectRole = "viewer"
	ProjectRoleParticipant ProjectRole = "participant"
	ProjectRoleEditor      ProjectRole = "editor"
	ProjectRoleManager     ProjectRole = "manager"
)

// ProjectListResult is the listing request specific struct
type ProjectListResult struct {
	Value []Project
}

// List the projects
func (s *ProjectsServiceClient) List(ctx context.Context) (*ProjectListResult, error) {
	req, err := s.client.newRequest(ctx, "GET", projectsBasePath, nil)
	if err != nil {
		return nil, err
	}

	err = s.client.setRequestListParameters(req, nil)
	if err != nil {
		return nil, err
	}

	var projectListResponse ProjectListResult
	_, err = s.client.do(ctx, req, &projectListResponse.Value)
	if err != nil {
		return nil, err
	}

	return &projectListResponse, nil
}

// Get a project
func (s *ProjectsServiceClient) Get(ctx context.Context, id int) (*Project, error) {
	path := fmt.Sprintf("%s/%d", projectsBasePath, id)

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var project Project
	_, err = s.client.do(ctx, req, &project)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// Create a project
func (s *ProjectsServiceClient) Create(ctx context.Context, project *Project) (*Project, error) {
	req, err := s.client.newRequest(ctx, "POST", projectsLegacyBasePath, project)
	if err != nil {
		return nil, err
	}

	var createdProject Project
	_, err = s.client.do(ctx, req, &createdProject)
	if err != nil {
		return nil, err
	}

	createdProject.Summary = project.Summary

	// update project after creation to establish the resource information.
	updatedProject, err := s.Update(ctx, *createdProject.ID, &createdProject)
	if err != nil {
		return nil, err
	}

	return updatedProject, nil
}

// Update a project
func (s *ProjectsServiceClient) Update(ctx context.Context, id int, project *Project) (*Project, error) {
	path := fmt.Sprintf("%s/%d", projectsLegacyBasePath, id)

	req, err := s.client.newRequest(ctx, "PUT", path, project)
	if err != nil {
		return nil, err
	}

	var updatedProject Project
	_, err = s.client.do(ctx, req, &updatedProject)
	if err != nil {
		return nil, err
	}

	return &updatedProject, nil
}

// Delete a project
func (s *ProjectsServiceClient) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("%s/%d", projectsLegacyBasePath, id)

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
