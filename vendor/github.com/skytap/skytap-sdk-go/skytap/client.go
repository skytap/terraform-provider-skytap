package skytap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	version   = "1.0.0"
	mediaType = "application/json"

	headerRequestID  = "X-Request-ID"
	headerRetryAfter = "Retry-After"

	defRetryAfter = 10
	defRetryCount = 6
)

// Client is a client to manage and configure the skytap cloud
type Client struct {
	// HTTP client to be used for communicating with the SkyTap SDK
	hc *http.Client

	// The base URL to be used when issuing requests
	BaseURL *url.URL

	// User agent used when issuing requests
	UserAgent string

	// Credentials provider to be used for authenticating with the API
	Credentials CredentialsProvider

	// Services used for communicating with the API
	Projects     ProjectsService
	Environments EnvironmentsService
	Templates    TemplatesService
	Networks     NetworksService
	VMs          VMsService

	retryAfter int
	retryCount int
}

// DefaultListParameters are the default pager settings
var DefaultListParameters = &ListParameters{
	Count:  intToPtr(100),
	Offset: intToPtr(0),
}

// ListParameters is a Client scoped common struct for listing
type ListParameters struct {
	// For paginated result sets, number of results to retrieve.
	Count *int

	// For paginated result sets, the offset of results to include.
	Offset *int

	// Filters
	Filters []ListFilter
}

// ListFilter is the struct for list filtering
type ListFilter struct {
	Name  *string
	Value *string
}

// ErrorResponse is the general purpose struct to hold error data
type ErrorResponse struct {
	error
	// HTTP response that caused this error
	Response *http.Response

	// RequestID returned from the API.
	RequestID *string

	// Error message
	Message *string `json:"error,omitempty"`

	// RetryAfter is sometimes returned by the server
	RetryAfter *int

	// RequiresRetry indicates whether a retry is required
	RequiresRetry bool
}

// Error returns a formatted error
func (r *ErrorResponse) Error() string {
	if r.RequestID != nil {
		return fmt.Sprintf("%v %v: %d (request %q) %v",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, *r.RequestID, *r.Message)
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

	baseURL, err := url.Parse(settings.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %v", baseURL)
	}

	client.BaseURL = baseURL
	client.UserAgent = settings.userAgent
	client.Credentials = settings.credentials

	client.Projects = &ProjectsServiceClient{&client}
	client.Environments = &EnvironmentsServiceClient{&client}
	client.Templates = &TemplatesServiceClient{&client}
	client.Networks = &NetworksServiceClient{&client}
	client.VMs = &VMsServiceClient{&client}

	client.retryAfter = defRetryAfter
	client.retryCount = defRetryCount

	return &client, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)
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
	var resp *http.Response
	var err error
	var makeRequest = true

	for i := 0; i < c.retryCount+1 && makeRequest; i++ {
		resp, err = c.hc.Do(req.WithContext(ctx))

		if err != nil {
			break
		}

		err = c.checkResponse(resp)

		if err == nil {
			readResponseBody(resp, v)
			makeRequest = false
		} else if err.(*ErrorResponse).RequiresRetry {
			seconds := *err.(*ErrorResponse).RetryAfter
			log.Printf("retrying after %d second(s)\n", seconds)
			time.Sleep(time.Duration(seconds) * time.Second)
		} else {
			makeRequest = false
		}
		resp.Body.Close()
	}

	return resp, err
}

func readResponseBody(resp *http.Response, v interface{}) error {
	var err error
	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}
	return err
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
func (c *Client) checkResponse(r *http.Response) error {
	if code := r.StatusCode; code >= http.StatusOK && code <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			errorResponse.Message = strToPtr(string(data))
		}
	}

	if requestID := r.Header.Get(headerRequestID); requestID != "" {
		errorResponse.RequestID = strToPtr(requestID)
	}

	if code := r.StatusCode; code == http.StatusLocked ||
		code == http.StatusTooManyRequests ||
		code >= http.StatusInternalServerError && code <= 599 {
		if retryAfter := r.Header.Get(headerRetryAfter); retryAfter != "" {
			val, err := strconv.Atoi(retryAfter)
			if err == nil {
				errorResponse.RetryAfter = intToPtr(val)
			} else {
				errorResponse.RetryAfter = intToPtr(c.retryAfter)
			}
		} else {
			errorResponse.RetryAfter = intToPtr(c.retryAfter)
		}
		errorResponse.RequiresRetry = true
	}

	return errorResponse
}
