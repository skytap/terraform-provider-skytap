package skytap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	defRetryCount = 60

	noRunStateCheck  RunStateCheckStatus = 0
	envRunStateCheck RunStateCheckStatus = 1
	vmRunStateCheck  RunStateCheckStatus = 2

	requestNotAsExpected = "request not as expected"
)

// RunStateCheckStatus value of the run check status used to determine checking of request and response.
type RunStateCheckStatus int

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
	Projects          ProjectsService
	Environments      EnvironmentsService
	Templates         TemplatesService
	Networks          NetworksService
	VMs               VMsService
	Interfaces        InterfacesService
	PublishedServices PublishedServicesService

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
	// HTTP response that caused this error
	Response *http.Response

	// RequestID returned from the API.
	RequestID *string

	// Error message
	Message *string `json:"error,omitempty"`

	// RetryAfter is sometimes returned by the server
	RetryAfter *int

	// RateLimited informs Skytap is rate limiting
	RateLimited *int
}

type environmentVMRunState struct {
	environmentID       *string
	vmID                *string
	adapterID           *string
	environment         []EnvironmentRunstate
	vm                  []VMRunstate
	diskIdentification  []DiskIdentification
	runStateCheckStatus RunStateCheckStatus
}

type responseComparator interface {
	compareResponse(ctx context.Context, c *Client, v interface{}, state *environmentVMRunState) (string, bool)
}

// Error returns a formatted error
func (r *ErrorResponse) Error() string {
	message := ""
	if r.Message != nil {
		message = *r.Message
	}
	ID := ""
	if r.RequestID != nil {
		ID = *r.RequestID
	}
	if r.RequestID != nil {
		return fmt.Sprintf("%v %v: %d (request %q) %v",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, ID, message)
	}

	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, message)
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
	client.Interfaces = &InterfacesServiceClient{&client}
	client.PublishedServices = &PublishedServicesServiceClient{&client}

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

func (c *Client) do(ctx context.Context, req *http.Request, v interface{}, state *environmentVMRunState, payload responseComparator) (*http.Response, error) {
	if req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodDelete {
		for i := 0; i < c.retryCount; i++ {
			err := c.checkResourceStateUntilSatisfied(ctx, req, state)
			if err != nil {
				return nil, err
			}
			resp, retry, err := c.requestPutPostDelete(ctx, req, state, payload, v)
			if !retry || err != nil {
				return resp, err
			}
		}
	}
	return c.request(ctx, req, v)
}

func (c *Client) request(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	err := logRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.hc.Do(req.WithContext(ctx))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		err := readResponseBody(resp, v)
		if err != nil {
			return nil, err
		}
	} else {
		return resp, c.buildErrorResponse(resp)
	}

	return resp, nil
}

func (c *Client) requestPutPostDelete(ctx context.Context, req *http.Request, state *environmentVMRunState, payload responseComparator, v interface{}) (*http.Response, bool, error) {
	var resp *http.Response
	err := logRequest(req)
	if err != nil {
		return nil, false, err
	}
	resp, err = c.hc.Do(req.WithContext(ctx))
	if err != nil {
		return nil, false, err
	}

	code := resp.StatusCode

	if code == http.StatusOK {
		err = readResponseBody(resp, v)
		if err != nil {
			return nil, false, err
		}
		if payload != nil {
			for i := 0; i < c.retryCount; i++ {
				if message, ok := payload.compareResponse(ctx, c, v, state); !ok {
					c.backoff("response check", fmt.Sprintf("%d", code), message, c.retryAfter)
				} else {
					return nil, false, nil
				}
			}
		}
		return nil, false, err
	}
	return c.handleError(resp, code)
}

func (c *Client) handleError(resp *http.Response, code int) (*http.Response, bool, error) {
	var errorSpecial *ErrorResponse
	errorSpecial = c.buildErrorResponse(resp).(*ErrorResponse)
	retryError := ""
	if code == http.StatusUnprocessableEntity {
		retryError = "StatusUnprocessableEntity"
		if !strings.Contains(*errorSpecial.Message, "busy") {
			return resp, false, errorSpecial
		}
	} else if code == http.StatusConflict {
		retryError = "StatusConflict"
	} else if code == http.StatusLocked {
		retryError = "StatusLocked"
	} else if code == http.StatusTooManyRequests {
		retryError = "StatusTooManyRequests"
	}
	if retryError != "" {
		seconds := c.retryAfter
		if errorSpecial.RetryAfter != nil {
			seconds = *errorSpecial.RetryAfter
		}
		c.backoff("response check", fmt.Sprintf("%d", code), retryError, seconds)
		return resp, true, nil
	}
	return resp, false, errorSpecial
}

func (c *Client) backoff(message string, code string, codeAsString string, snooze int) {
	log.Printf("[INFO] SDK %s (%s:%s). Retrying after %d second(s)\n", message, code, codeAsString, snooze)
	time.Sleep(time.Duration(snooze) * time.Second)
}

func readResponseBody(resp *http.Response, v interface{}) error {
	var err error
	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
		if err != nil {
			log.Printf("[ERROR] SDK response payload decoding: (%s)", err.Error())
		}
		err = resp.Body.Close()
	}
	return err
}

func logRequest(req *http.Request) error {
	log.Printf("[DEBUG] SDK request (%s), URL (%s), agent (%s)\n", req.Method, req.URL.String(), req.UserAgent())
	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
		log.Printf("[DEBUG] SDK request body (%s)\n", strings.TrimSpace(string(body)))
	}
	return nil
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

func (c *Client) buildErrorResponse(r *http.Response) error {
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		errorResponse.Message = strToPtr(string(data))
		log.Printf("[INFO] SDK response error: (%s)", *errorResponse.Message)
	}

	if requestID := r.Header.Get(headerRequestID); requestID != "" {
		errorResponse.RequestID = strToPtr(requestID)
	}

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

	return errorResponse
}

func (c *Client) checkResourceStateUntilSatisfied(ctx context.Context, req *http.Request, state *environmentVMRunState) error {
	runStateCheck := c.requiresChecking(state)
	if runStateCheck > noRunStateCheck {
		for i := 0; i < c.retryCount; i++ {
			var ok bool
			var err error
			if runStateCheck == envRunStateCheck {
				ok, err = c.getEnvironmentRunState(ctx, state.environmentID, state.environment)
			} else {
				ok, err = c.getVMRunState(ctx, state.environmentID, state.vmID, state.vm)
			}
			if err != nil || ok {
				return err
			}
			c.backoff("pre-check loop", "", "", c.retryAfter)
		}
		return errors.New("timeout waiting for state")
	}
	return nil
}

func (c *Client) requiresChecking(state *environmentVMRunState) RunStateCheckStatus {
	if state == nil {
		return noRunStateCheck
	}
	return state.runStateCheckStatus
}

func (c *Client) getEnvironmentRunState(ctx context.Context, id *string, states []EnvironmentRunstate) (bool, error) {
	env, err := c.Environments.Get(ctx, *id)
	if err != nil {
		return false, err
	}
	if env.Runstate == nil {
		return false, errors.New("environment run state not set")
	}
	ok := c.containsEnvironmentRunState(env.Runstate, states)
	log.Printf("[DEBUG] SDK run state of environment (%s) and require: (%s).\n",
		*env.Runstate,
		c.environmentsRunStatesToString(states))
	return ok, nil
}

func (c *Client) containsEnvironmentRunState(currentState *EnvironmentRunstate, possibleStates []EnvironmentRunstate) bool {
	for _, v := range possibleStates {
		if v == *currentState {
			return true
		}
	}
	return false
}

func (c *Client) environmentsRunStatesToString(possibleStates []EnvironmentRunstate) string {
	var items []string
	for _, v := range possibleStates {
		items = append(items, string(v))
	}
	return strings.Join(items, ", ")
}

func (c *Client) getVMRunState(ctx context.Context, environmentID *string, vmID *string, states []VMRunstate) (bool, error) {
	vm, err := c.VMs.Get(ctx, *environmentID, *vmID)
	if err != nil {
		return false, err
	}
	if vm.Runstate == nil {
		return false, errors.New("vm run state not set")
	}
	ok := c.containsVMRunState(vm.Runstate, states)
	log.Printf("[INFO] SDK run state of vm (%s) and require: (%s).\n",
		*vm.Runstate,
		c.vMRunStatesToString(states))
	return ok, nil
}

func (c *Client) containsVMRunState(currentState *VMRunstate, possibleStates []VMRunstate) bool {
	for _, v := range possibleStates {
		if v == *currentState {
			return true
		}
	}
	return false
}

func (c *Client) vMRunStatesToString(possibleStates []VMRunstate) string {
	var items []string
	for _, v := range possibleStates {
		items = append(items, string(v))
	}
	return strings.Join(items, ", ")
}

func envRunStateNotBusy(environmentID string) *environmentVMRunState {
	return &environmentVMRunState{
		environmentID: strToPtr(environmentID),
		environment: []EnvironmentRunstate{
			EnvironmentRunstateRunning,
			EnvironmentRunstateStopped,
			EnvironmentRunstateSuspended,
			EnvironmentRunstateHalted},
		runStateCheckStatus: envRunStateCheck,
	}
}

func envRunStateNotBusyWithVM(environmentID string, vmID string) *environmentVMRunState {
	state := envRunStateNotBusy(environmentID)
	state.vmID = strToPtr(vmID)
	return state
}

func vmRunStateNotBusy(environmentID string, vmID string) *environmentVMRunState {
	return &environmentVMRunState{
		environmentID: strToPtr(environmentID),
		vmID:          strToPtr(vmID),
		vm: []VMRunstate{
			VMRunstateStopped,
			VMRunstateHalted,
			VMRunstateReset,
			VMRunstateRunning,
			VMRunstateSuspended},
		runStateCheckStatus: vmRunStateCheck,
	}
}

func vmRequestRunStateStopped(environmentID string, vmID string) *environmentVMRunState {
	state := vmRunStateNotBusy(environmentID, vmID)
	state.vm = []VMRunstate{VMRunstateStopped}
	return state
}

func vmRunStateNotBusyWithAdapter(environmentID string, vmID string, adapterID string) *environmentVMRunState {
	state := vmRunStateNotBusy(environmentID, vmID)
	state.adapterID = strToPtr(adapterID)
	return state
}
