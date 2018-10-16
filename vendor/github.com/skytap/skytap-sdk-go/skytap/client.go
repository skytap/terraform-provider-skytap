package skytap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	version   = "1.0.0"
	mediaType = "application/json"

	headerRequestId = "X-Request-Id"
)

// Client is a client to manage and configure the skytap cloud
type Client struct {
	// HTTP client to be used for communicating with the SkyTap SDK
	hc *http.Client

	// The base URL to be used when issuing requests
	BaseUrl *url.URL

	// User agent used when issuing requests
	UserAgent string

	// Credentials provider to be used for authenticating with the API
	Credentials CredentialsProvider

	// Services used for communicating with the API
	Projects     ProjectsService
	Environments EnvironmentsService
}

var DefaultListParameters = &ListParameters{
	Count:  IntPtr(100),
	Offset: IntPtr(0),
}

// Client scoped common structs
type ListParameters struct {
	// For paginated result sets, number of results to retrieve.
	Count *int

	// For paginated result sets, the offset of results to include.
	Offset *int

	// Filters
	Filters []ListFilter
}

type ListFilter struct {
	Name  *string
	Value *string
}

type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	// RequestId returned from the API.
	RequestId *string

	// Error message
	Message *string `json:"error,omitempty"`
}

func (r *ErrorResponse) Error() string {
	if r.RequestId != nil {
		return fmt.Sprintf("%v %v: %d (request %q) %v",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, *r.RequestId, *r.Message)
	}

	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, *r.Message)
}

// NewClient creates a Skytab cloud client
func NewClient(settings Settings) (*Client, error) {
	if err := settings.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate client config: %v", err)
	}

	client := Client{
		hc: http.DefaultClient,
	}

	baseUrl, err := url.Parse(settings.baseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %v", baseUrl)
	}

	client.BaseUrl = baseUrl
	client.UserAgent = settings.userAgent
	client.Credentials = settings.credentials

	client.Projects = &ProjectsServiceClient{&client}
	client.Environments = &EnvironmentsServiceClient{&client}

	return &client, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.BaseUrl.ResolveReference(rel)
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", mediaType)
	}
	req.Header.Set("Accept", mediaType)
	req.Header.Set("User-Agent", c.UserAgent)

	// Retrieve the authentication/authorization header from the clients credential provider
	auth, err := c.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, err
	}

	if auth != "" {
		req.Header.Set("Authorization", auth)
	}

	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.hc.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = checkResponse(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				return nil, err
			}
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				return nil, err
			}
		}
	}

	return resp, err
}

func (c *Client) setRequestListParameters(req *http.Request, params *ListParameters) error {
	if params == nil {
		params = DefaultListParameters
	}

	q := req.URL.Query()

	if v := params.Count; v != nil {
		q.Add("count", strconv.Itoa(*v))
	}
	if v := params.Offset; v != nil {
		q.Add("offset", strconv.Itoa(*v))
	}

	if v := params.Filters; v != nil && len(v) > 0 {
		var filters []string
		for _, f := range v {
			if f.Name != nil && f.Value != nil {
				filters = append(filters, fmt.Sprintf("%s:%s", *f.Name, *f.Value))
			}
		}

		q.Add("query", strings.Join(filters, ","))
	}

	req.URL.RawQuery = q.Encode()

	return nil
}

// checkResponse checks the API response for errors, and returns them if present. A response is considered an
// error if it has a status code outside the 200 range. API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse.
func checkResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			errorResponse.Message = StringPtr(string(data))
		}
	}

	if requestId := r.Header.Get(headerRequestId); requestId != "" {
		errorResponse.RequestId = StringPtr(requestId)
	}

	return errorResponse
}
