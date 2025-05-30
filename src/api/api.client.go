// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oapi-codegen/runtime"
)

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// GetBots request
	GetBots(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetBotsBotId request
	GetBotsBotId(ctx context.Context, botId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostBotsBotIdMine request
	PostBotsBotIdMine(ctx context.Context, botId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostBotsBotIdMoveWithBody request with any body
	PostBotsBotIdMoveWithBody(ctx context.Context, botId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostBotsBotIdMove(ctx context.Context, botId string, body PostBotsBotIdMoveJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostBotsBotIdNewBotWithBody request with any body
	PostBotsBotIdNewBotWithBody(ctx context.Context, botId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostBotsBotIdNewBot(ctx context.Context, botId string, body PostBotsBotIdNewBotJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostInit request
	PostInit(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// Login request
	Login(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetMines request
	GetMines(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// NewUser request
	NewUser(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) GetBots(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetBotsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetBotsBotId(ctx context.Context, botId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetBotsBotIdRequest(c.Server, botId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostBotsBotIdMine(ctx context.Context, botId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostBotsBotIdMineRequest(c.Server, botId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostBotsBotIdMoveWithBody(ctx context.Context, botId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostBotsBotIdMoveRequestWithBody(c.Server, botId, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostBotsBotIdMove(ctx context.Context, botId string, body PostBotsBotIdMoveJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostBotsBotIdMoveRequest(c.Server, botId, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostBotsBotIdNewBotWithBody(ctx context.Context, botId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostBotsBotIdNewBotRequestWithBody(c.Server, botId, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostBotsBotIdNewBot(ctx context.Context, botId string, body PostBotsBotIdNewBotJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostBotsBotIdNewBotRequest(c.Server, botId, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostInit(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostInitRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) Login(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewLoginRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetMines(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetMinesRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) NewUser(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewNewUserRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewGetBotsRequest generates requests for GetBots
func NewGetBotsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/bots")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetBotsBotIdRequest generates requests for GetBotsBotId
func NewGetBotsBotIdRequest(server string, botId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "botId", runtime.ParamLocationPath, botId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/bots/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPostBotsBotIdMineRequest generates requests for PostBotsBotIdMine
func NewPostBotsBotIdMineRequest(server string, botId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "botId", runtime.ParamLocationPath, botId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/bots/%s/mine", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPostBotsBotIdMoveRequest calls the generic PostBotsBotIdMove builder with application/json body
func NewPostBotsBotIdMoveRequest(server string, botId string, body PostBotsBotIdMoveJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostBotsBotIdMoveRequestWithBody(server, botId, "application/json", bodyReader)
}

// NewPostBotsBotIdMoveRequestWithBody generates requests for PostBotsBotIdMove with any type of body
func NewPostBotsBotIdMoveRequestWithBody(server string, botId string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "botId", runtime.ParamLocationPath, botId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/bots/%s/move", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewPostBotsBotIdNewBotRequest calls the generic PostBotsBotIdNewBot builder with application/json body
func NewPostBotsBotIdNewBotRequest(server string, botId string, body PostBotsBotIdNewBotJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostBotsBotIdNewBotRequestWithBody(server, botId, "application/json", bodyReader)
}

// NewPostBotsBotIdNewBotRequestWithBody generates requests for PostBotsBotIdNewBot with any type of body
func NewPostBotsBotIdNewBotRequestWithBody(server string, botId string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "botId", runtime.ParamLocationPath, botId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/bots/%s/newBot", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewPostInitRequest generates requests for PostInit
func NewPostInitRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/init")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewLoginRequest generates requests for Login
func NewLoginRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/login")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetMinesRequest generates requests for GetMines
func NewGetMinesRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/mines")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewNewUserRequest generates requests for NewUser
func NewNewUserRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/newUser")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// GetBotsWithResponse request
	GetBotsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetBotsResponse, error)

	// GetBotsBotIdWithResponse request
	GetBotsBotIdWithResponse(ctx context.Context, botId string, reqEditors ...RequestEditorFn) (*GetBotsBotIdResponse, error)

	// PostBotsBotIdMineWithResponse request
	PostBotsBotIdMineWithResponse(ctx context.Context, botId string, reqEditors ...RequestEditorFn) (*PostBotsBotIdMineResponse, error)

	// PostBotsBotIdMoveWithBodyWithResponse request with any body
	PostBotsBotIdMoveWithBodyWithResponse(ctx context.Context, botId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostBotsBotIdMoveResponse, error)

	PostBotsBotIdMoveWithResponse(ctx context.Context, botId string, body PostBotsBotIdMoveJSONRequestBody, reqEditors ...RequestEditorFn) (*PostBotsBotIdMoveResponse, error)

	// PostBotsBotIdNewBotWithBodyWithResponse request with any body
	PostBotsBotIdNewBotWithBodyWithResponse(ctx context.Context, botId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostBotsBotIdNewBotResponse, error)

	PostBotsBotIdNewBotWithResponse(ctx context.Context, botId string, body PostBotsBotIdNewBotJSONRequestBody, reqEditors ...RequestEditorFn) (*PostBotsBotIdNewBotResponse, error)

	// PostInitWithResponse request
	PostInitWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*PostInitResponse, error)

	// LoginWithResponse request
	LoginWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*LoginResponse, error)

	// GetMinesWithResponse request
	GetMinesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetMinesResponse, error)

	// NewUserWithResponse request
	NewUserWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*NewUserResponse, error)
}

type GetBotsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]Bot
}

// Status returns HTTPResponse.Status
func (r GetBotsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetBotsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetBotsBotIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Bot
}

// Status returns HTTPResponse.Status
func (r GetBotsBotIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetBotsBotIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostBotsBotIdMineResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Bot
}

// Status returns HTTPResponse.Status
func (r PostBotsBotIdMineResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostBotsBotIdMineResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostBotsBotIdMoveResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Bot
}

// Status returns HTTPResponse.Status
func (r PostBotsBotIdMoveResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostBotsBotIdMoveResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostBotsBotIdNewBotResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Bot
}

// Status returns HTTPResponse.Status
func (r PostBotsBotIdNewBotResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostBotsBotIdNewBotResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostInitResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Bots  *[]Bot         `json:"bots,omitempty"`
		Mines *[]Coordinates `json:"mines,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r PostInitResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostInitResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type LoginResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r LoginResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r LoginResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetMinesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]Coordinates
}

// Status returns HTTPResponse.Status
func (r GetMinesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetMinesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type NewUserResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Message *string `json:"message,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r NewUserResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r NewUserResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// GetBotsWithResponse request returning *GetBotsResponse
func (c *ClientWithResponses) GetBotsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetBotsResponse, error) {
	rsp, err := c.GetBots(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetBotsResponse(rsp)
}

// GetBotsBotIdWithResponse request returning *GetBotsBotIdResponse
func (c *ClientWithResponses) GetBotsBotIdWithResponse(ctx context.Context, botId string, reqEditors ...RequestEditorFn) (*GetBotsBotIdResponse, error) {
	rsp, err := c.GetBotsBotId(ctx, botId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetBotsBotIdResponse(rsp)
}

// PostBotsBotIdMineWithResponse request returning *PostBotsBotIdMineResponse
func (c *ClientWithResponses) PostBotsBotIdMineWithResponse(ctx context.Context, botId string, reqEditors ...RequestEditorFn) (*PostBotsBotIdMineResponse, error) {
	rsp, err := c.PostBotsBotIdMine(ctx, botId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostBotsBotIdMineResponse(rsp)
}

// PostBotsBotIdMoveWithBodyWithResponse request with arbitrary body returning *PostBotsBotIdMoveResponse
func (c *ClientWithResponses) PostBotsBotIdMoveWithBodyWithResponse(ctx context.Context, botId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostBotsBotIdMoveResponse, error) {
	rsp, err := c.PostBotsBotIdMoveWithBody(ctx, botId, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostBotsBotIdMoveResponse(rsp)
}

func (c *ClientWithResponses) PostBotsBotIdMoveWithResponse(ctx context.Context, botId string, body PostBotsBotIdMoveJSONRequestBody, reqEditors ...RequestEditorFn) (*PostBotsBotIdMoveResponse, error) {
	rsp, err := c.PostBotsBotIdMove(ctx, botId, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostBotsBotIdMoveResponse(rsp)
}

// PostBotsBotIdNewBotWithBodyWithResponse request with arbitrary body returning *PostBotsBotIdNewBotResponse
func (c *ClientWithResponses) PostBotsBotIdNewBotWithBodyWithResponse(ctx context.Context, botId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostBotsBotIdNewBotResponse, error) {
	rsp, err := c.PostBotsBotIdNewBotWithBody(ctx, botId, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostBotsBotIdNewBotResponse(rsp)
}

func (c *ClientWithResponses) PostBotsBotIdNewBotWithResponse(ctx context.Context, botId string, body PostBotsBotIdNewBotJSONRequestBody, reqEditors ...RequestEditorFn) (*PostBotsBotIdNewBotResponse, error) {
	rsp, err := c.PostBotsBotIdNewBot(ctx, botId, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostBotsBotIdNewBotResponse(rsp)
}

// PostInitWithResponse request returning *PostInitResponse
func (c *ClientWithResponses) PostInitWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*PostInitResponse, error) {
	rsp, err := c.PostInit(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostInitResponse(rsp)
}

// LoginWithResponse request returning *LoginResponse
func (c *ClientWithResponses) LoginWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*LoginResponse, error) {
	rsp, err := c.Login(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseLoginResponse(rsp)
}

// GetMinesWithResponse request returning *GetMinesResponse
func (c *ClientWithResponses) GetMinesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetMinesResponse, error) {
	rsp, err := c.GetMines(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetMinesResponse(rsp)
}

// NewUserWithResponse request returning *NewUserResponse
func (c *ClientWithResponses) NewUserWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*NewUserResponse, error) {
	rsp, err := c.NewUser(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseNewUserResponse(rsp)
}

// ParseGetBotsResponse parses an HTTP response from a GetBotsWithResponse call
func ParseGetBotsResponse(rsp *http.Response) (*GetBotsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetBotsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []Bot
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetBotsBotIdResponse parses an HTTP response from a GetBotsBotIdWithResponse call
func ParseGetBotsBotIdResponse(rsp *http.Response) (*GetBotsBotIdResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetBotsBotIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Bot
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParsePostBotsBotIdMineResponse parses an HTTP response from a PostBotsBotIdMineWithResponse call
func ParsePostBotsBotIdMineResponse(rsp *http.Response) (*PostBotsBotIdMineResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostBotsBotIdMineResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Bot
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParsePostBotsBotIdMoveResponse parses an HTTP response from a PostBotsBotIdMoveWithResponse call
func ParsePostBotsBotIdMoveResponse(rsp *http.Response) (*PostBotsBotIdMoveResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostBotsBotIdMoveResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Bot
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParsePostBotsBotIdNewBotResponse parses an HTTP response from a PostBotsBotIdNewBotWithResponse call
func ParsePostBotsBotIdNewBotResponse(rsp *http.Response) (*PostBotsBotIdNewBotResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostBotsBotIdNewBotResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Bot
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParsePostInitResponse parses an HTTP response from a PostInitWithResponse call
func ParsePostInitResponse(rsp *http.Response) (*PostInitResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostInitResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Bots  *[]Bot         `json:"bots,omitempty"`
			Mines *[]Coordinates `json:"mines,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseLoginResponse parses an HTTP response from a LoginWithResponse call
func ParseLoginResponse(rsp *http.Response) (*LoginResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &LoginResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseGetMinesResponse parses an HTTP response from a GetMinesWithResponse call
func ParseGetMinesResponse(rsp *http.Response) (*GetMinesResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetMinesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []Coordinates
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseNewUserResponse parses an HTTP response from a NewUserWithResponse call
func ParseNewUserResponse(rsp *http.Response) (*NewUserResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &NewUserResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Message *string `json:"message,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}
