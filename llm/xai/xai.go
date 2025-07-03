package xai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/maksymenkoml/lingoose/llm/openai"
	"github.com/maksymenkoml/lingoose/thread"
	goopenai "github.com/sashabaranov/go-openai"
)

const (
	xaiAPIEndpoint = "https://api.x.ai/v1"
)

type XAI struct {
	*openai.OpenAI
	searchMode   string // "auto", "on", "off"
	currentModel Model
}

func New() *XAI {
	customConfig := goopenai.DefaultConfig(os.Getenv("XAI_API_KEY"))
	customConfig.BaseURL = xaiAPIEndpoint

	// Create XAI instance first
	xai := &XAI{
		searchMode:   "off",       // default to off
		currentModel: Grok3Latest, // Default model
	}

	// Create custom HTTP client with our transport
	httpClient := &http.Client{
		Transport: newCustomTransport(xai.searchMode),
	}
	customConfig.HTTPClient = httpClient

	customClient := goopenai.NewClientWithConfig(customConfig)
	openaillm := openai.New().WithClient(customClient)
	openaillm.Name = "xai"

	xai.OpenAI = openaillm
	return xai
}

func (x *XAI) WithModel(model Model) *XAI {
	x.currentModel = model
	x.OpenAI.WithModel(openai.Model(model))
	return x
}

// WithSearchMode sets the search mode for the X.AI instance
// mode can be "auto", "on", or "off"
func (x *XAI) WithSearchMode(mode string) *XAI {
	x.searchMode = mode
	x.updateHTTPTransport()
	return x
}

// updateHTTPTransport updates the HTTP transport with current search settings
func (x *XAI) updateHTTPTransport() {
	// We need to access the underlying HTTP client and update its transport
	// This is a bit tricky since go-openai doesn't expose the HTTP client directly
	// We'll use reflection or recreation approach

	// For now, let's recreate the client with new transport
	customConfig := goopenai.DefaultConfig(os.Getenv("XAI_API_KEY"))
	customConfig.BaseURL = xaiAPIEndpoint

	httpClient := &http.Client{
		Transport: newCustomTransport(x.searchMode),
	}
	customConfig.HTTPClient = httpClient

	customClient := goopenai.NewClientWithConfig(customConfig)
	x.OpenAI.WithClient(customClient)
}

// WithLiveSearch enables live search functionality for the X.AI instance (deprecated, use WithSearchMode)
func (x *XAI) WithLiveSearch(enable bool) *XAI {
	if enable {
		return x.WithSearchMode("auto")
	}
	return x.WithSearchMode("off")
}

// Generate generates a completion using X.AI with optional search
func (x *XAI) Generate(ctx context.Context, t *thread.Thread) error {
	// CustomTransport automatically adds search_parameters based on searchMode
	return x.OpenAI.Generate(ctx, t)
}

// CustomTransport is a custom HTTP transport that modifies requests to add search_parameters
type CustomTransport struct {
	base       http.RoundTripper
	searchMode string
}

func newCustomTransport(searchMode string) *CustomTransport {
	return &CustomTransport{
		base:       http.DefaultTransport,
		searchMode: searchMode,
	}
}

func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only modify requests to xAI API
	if !strings.Contains(req.URL.String(), "api.x.ai") {
		return c.base.RoundTrip(req)
	}

	// Only modify POST requests to chat/completions
	if req.Method != "POST" || !strings.Contains(req.URL.Path, "chat/completions") {
		return c.base.RoundTrip(req)
	}

	// Read the original request body
	if req.Body == nil {
		return c.base.RoundTrip(req)
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return c.base.RoundTrip(req)
	}
	req.Body.Close()

	// Parse the JSON request
	var requestData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
		// If we can't parse, just pass through
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return c.base.RoundTrip(req)
	}
	// Add search_parameters if search mode is not "off"
	if c.searchMode != "off" {
		searchParams := map[string]interface{}{
			"mode": c.searchMode,
		}
		requestData["search_parameters"] = searchParams
	}

	// Marshal back to JSON
	modifiedBody, err := json.Marshal(requestData)
	if err != nil {
		// If we can't marshal, use original body
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return c.base.RoundTrip(req)
	}

	// Create new request with modified body
	req.Body = io.NopCloser(bytes.NewBuffer(modifiedBody))
	req.ContentLength = int64(len(modifiedBody))

	return c.base.RoundTrip(req)
}
