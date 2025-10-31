package domain

import "time"

// ComposeBodyMode describes how request body is constructed.
// Supported: "raw", "json", "form", "multipart".
type ComposeBodyMode string

const (
	ComposeBodyRaw       ComposeBodyMode = "raw"
	ComposeBodyJSON      ComposeBodyMode = "json"
	ComposeBodyForm      ComposeBodyMode = "form"      // application/x-www-form-urlencoded
	ComposeBodyMultipart ComposeBodyMode = "multipart" // multipart/form-data
)

type ComposeHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ComposeQueryParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ComposeFormField represents a key/value pair for x-www-form-urlencoded body.
type ComposeFormField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ComposeMultipartPart represents a single multipart/form-data part.
// If IsFile is true, then FileName and Base64 must be provided.
// Otherwise Value is used as text field.
type ComposeMultipartPart struct {
	Name        string `json:"name"`
	IsFile      bool   `json:"isFile"`
	FileName    string `json:"fileName,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Base64      string `json:"base64,omitempty"`
	Value       string `json:"value,omitempty"`
}

// ComposeAuth represents authentication settings for a request.
// Type: "none", "basic", "bearer", "apiKey" (apiKey: header="X-Api-Key").
type ComposeAuth struct {
	Type         string `json:"type"`
	Username     string `json:"username,omitempty"`     // basic
	Password     string `json:"password,omitempty"`     // basic
	BearerToken  string `json:"bearerToken,omitempty"`  // bearer
	APIKey       string `json:"apiKey,omitempty"`       // apiKey
	APIKeyHeader string `json:"apiKeyHeader,omitempty"` // default X-Api-Key
}

// ComposeBody describes body mode and content.
type ComposeBody struct {
	Mode      ComposeBodyMode        `json:"mode"`
	Raw       string                 `json:"raw,omitempty"`
	JSON      string                 `json:"json,omitempty"`
	Form      []ComposeFormField     `json:"form,omitempty"`
	Multipart []ComposeMultipartPart `json:"multipart,omitempty"`
}

// ComposeRequestTemplate describes a saved request template in library.
type ComposeRequestTemplate struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Method    string              `json:"method"`
	URL       string              `json:"url"`
	Headers   []ComposeHeader     `json:"headers,omitempty"`
	Query     []ComposeQueryParam `json:"query,omitempty"`
	Body      ComposeBody         `json:"body"`
	Auth      *ComposeAuth        `json:"auth,omitempty"`
	CreatedAt time.Time           `json:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt"`
}

// ComposeFolder is a tree node that can contain requests and subfolders.
type ComposeFolder struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Requests []string        `json:"requests"` // request IDs
	Folders  []ComposeFolder `json:"folders"`
}

// ComposeCollection groups folders/requests under a named collection.
type ComposeCollection struct {
	ID   string        `json:"id"`
	Name string        `json:"name"`
	Root ComposeFolder `json:"root"`
}

// ComposeLibrary contains all collections and the map of request templates.
// RequestsByID acts as a storage for request templates referenced by folders.
type ComposeLibrary struct {
	Collections  []ComposeCollection               `json:"collections"`
	RequestsByID map[string]ComposeRequestTemplate `json:"requestsById"`
}

// ComposeHistoryEntry represents a single send event for quick access.
type ComposeHistoryEntry struct {
	ID       string    `json:"id"`
	Template string    `json:"templateId,omitempty"`
	Method   string    `json:"method"`
	URL      string    `json:"url"`
	Status   int       `json:"status"`
	When     time.Time `json:"when"`
}
