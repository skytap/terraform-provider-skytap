package skytap

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)

// Default URL paths
const (
	labelCategoryBasePath = "/v2/label_categories"
)

// LabelCategory model a label category entity
type LabelCategory struct {
	ID          *int    `json:"id,omitempty,string"`
	Name        *string `json:"name,omitempty"`
	SingleValue *bool   `json:"single_value,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
}

// LabelCategoryService provides the entity contract
type LabelCategoryService interface {
	List(ctx context.Context) ([]LabelCategory, error)
	Get(ctx context.Context, id int) (*LabelCategory, error)
	Create(ctx context.Context, category *LabelCategory) (*LabelCategory, error)
	Delete(ctx context.Context, id int) error
}

// LabelCategoryClient provides the entity implementation
type LabelCategoryClient struct {
	client *Client
}

type enableUpdate struct {
	Enabled bool `json:"enabled"`
}

type labelCategoryError struct {
	Error string `json:"error,omitempty"`
	URL   string `json:"url,omitempty"`
}

// List Label category
func (s *LabelCategoryClient) List(ctx context.Context) ([]LabelCategory, error) {
	req, err := s.client.newRequest(ctx, "GET", labelCategoryBasePath, nil)
	if err != nil {
		return nil, err
	}

	err = s.client.setRequestListParameters(req, nil)
	if err != nil {
		return nil, err
	}

	var labelCategoryList []LabelCategory
	_, err = s.client.do(ctx, req, &labelCategoryList, nil, nil)
	if err != nil {
		return nil, err
	}

	return labelCategoryList, nil
}

// Get retrieves a label category
func (s *LabelCategoryClient) Get(ctx context.Context, id int) (*LabelCategory, error) {
	path := fmt.Sprintf("%s/%d", labelCategoryBasePath, id)
	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var labelCategory LabelCategory
	_, err = s.client.do(ctx, req, &labelCategory, nil, nil)
	if err != nil {
		return nil, err
	}

	return &labelCategory, nil
}

// Get creates a label category
func (s *LabelCategoryClient) Create(ctx context.Context, category *LabelCategory) (*LabelCategory, error) {
	req, err := s.client.newRequest(ctx, "POST", labelCategoryBasePath, category)
	if err != nil {
		return nil, err
	}

	var createdCategory LabelCategory
	var labelCategoryError labelCategoryError
	// The creation of a label category that is disable return a body error
	// the id of the label category by name can be obtained by filtering the
	// category list, in this case we use a native client to retrieve and process
	// the error.
	resp, err := s.client.hc.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case 200:
		readResponseBody(resp, &createdCategory)
	case 409:
		// In this case the category is disabled, we enable the exiting category
		log.Println("[DEBUG] SDK restoring a delete label category.")
		readResponseBody(resp, &labelCategoryError)

		if strings.HasPrefix(strings.ToLower(labelCategoryError.Error), "validation failed:") {
			return nil, fmt.Errorf(`Error creating label category with name (%s): %s`, *category.Name, labelCategoryError.Error)
		}

		labelCategoryID, err := labelCategoryResourceIdFromURL(labelCategoryError.URL)
		path := fmt.Sprintf("%s/%d.json", labelCategoryBasePath, labelCategoryID)
		enabled := enableUpdate{true}
		updateReq, err := s.client.newRequest(ctx, "PUT", path, enabled)
		if err != nil {
			return nil, err
		}
		_, err = s.client.do(ctx, updateReq, &createdCategory, nil, nil)
		if err != nil {
			return nil, err
		}

		if *category.SingleValue != *createdCategory.SingleValue {
			// try to rollback the existing resource
			s.Delete(ctx, labelCategoryID)
			return nil, fmt.Errorf("The label category with id: %d can not be created with this single value property"+
				" as it is recreated from a existing label category.", labelCategoryID)
		}

		if err != nil {
			return nil, err
		}

		// Update request do not contain the id
		createdCategory.ID = &labelCategoryID
	default:
		return nil, fmt.Errorf("Unexpected status code %d creating label category", resp.StatusCode)
	}

	return &createdCategory, nil
}

// Delete a label category
func (s *LabelCategoryClient) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("%s/%d.json", labelCategoryBasePath, id)

	softDeleteUpdate := enableUpdate{false}
	req, err := s.client.newRequest(ctx, "PUT", path, softDeleteUpdate)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// this helper methods help to extract the id of the category from the URL, as is not
// structured in the body of the error response.
func labelCategoryResourceIdFromURL(labelCategoryResource string) (int, error) {
	parsedURL, err := url.Parse(labelCategoryResource)
	if err != nil {
		return 0, err
	}
	var ps string
	path := parsedURL.Path
	for i := 1; i < len(path); i++ {
		if path[i] == '/' {
			ps = path[i+1:]
		}
	}
	return strconv.Atoi(ps)
}
