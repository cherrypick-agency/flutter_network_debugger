package httpapi

import (
	"reflect"
	"testing"

	"network-debugger/internal/domain"
)

func TestDtoToDomainTemplate(t *testing.T) {
	tests := []struct {
		name string
		dto  composeTemplateDTO
		want domain.ComposeRequestTemplate
	}{
		{
			name: "basic template",
			dto: composeTemplateDTO{
				ID:     "test-1",
				Name:   "Test Request",
				Method: "post",
				URL:    "https://api.example.com/test",
			},
			want: domain.ComposeRequestTemplate{
				ID:      "test-1",
				Name:    "Test Request",
				Method:  "POST",
				URL:     "https://api.example.com/test",
				Headers: []domain.ComposeHeader{},
				Query:   []domain.ComposeQueryParam{},
				Body:    domain.ComposeBody{Mode: domain.ComposeBodyMode("")},
			},
		},
		{
			name: "with headers",
			dto: composeTemplateDTO{
				ID:     "test-2",
				Name:   "With Headers",
				Method: "get",
				URL:    "https://example.com",
				Headers: []composeHeaderDTO{
					{Key: "Content-Type", Value: "application/json"},
					{Key: "Authorization", Value: "Bearer token"},
				},
			},
			want: domain.ComposeRequestTemplate{
				ID:     "test-2",
				Name:   "With Headers",
				Method: "GET",
				URL:    "https://example.com",
				Headers: []domain.ComposeHeader{
					{Key: "Content-Type", Value: "application/json"},
					{Key: "Authorization", Value: "Bearer token"},
				},
				Query: []domain.ComposeQueryParam{},
				Body:  domain.ComposeBody{Mode: domain.ComposeBodyMode("")},
			},
		},
		{
			name: "with query params",
			dto: composeTemplateDTO{
				ID:     "test-3",
				Name:   "With Query",
				Method: "GET",
				URL:    "https://example.com/search",
				Query: []composeQueryDTO{
					{Key: "q", Value: "test"},
					{Key: "page", Value: "1"},
				},
			},
			want: domain.ComposeRequestTemplate{
				ID:      "test-3",
				Name:    "With Query",
				Method:  "GET",
				URL:     "https://example.com/search",
				Headers: []domain.ComposeHeader{},
				Query: []domain.ComposeQueryParam{
					{Key: "q", Value: "test"},
					{Key: "page", Value: "1"},
				},
				Body: domain.ComposeBody{Mode: domain.ComposeBodyMode("")},
			},
		},
		{
			name: "with auth bearer",
			dto: composeTemplateDTO{
				ID:     "test-4",
				Name:   "With Auth",
				Method: "POST",
				URL:    "https://api.example.com",
				Auth: &composeAuthDTO{
					Type:        "bearer",
					BearerToken: "my-token-123",
				},
			},
			want: domain.ComposeRequestTemplate{
				ID:      "test-4",
				Name:    "With Auth",
				Method:  "POST",
				URL:     "https://api.example.com",
				Headers: []domain.ComposeHeader{},
				Query:   []domain.ComposeQueryParam{},
				Auth: &domain.ComposeAuth{
					Type:        "bearer",
					BearerToken: "my-token-123",
				},
				Body: domain.ComposeBody{Mode: domain.ComposeBodyMode("")},
			},
		},
		{
			name: "with auth basic",
			dto: composeTemplateDTO{
				ID:     "test-5",
				Name:   "With Basic Auth",
				Method: "GET",
				URL:    "https://example.com",
				Auth: &composeAuthDTO{
					Type:     "basic",
					Username: "user",
					Password: "pass",
				},
			},
			want: domain.ComposeRequestTemplate{
				ID:      "test-5",
				Name:    "With Basic Auth",
				Method:  "GET",
				URL:     "https://example.com",
				Headers: []domain.ComposeHeader{},
				Query:   []domain.ComposeQueryParam{},
				Auth: &domain.ComposeAuth{
					Type:     "basic",
					Username: "user",
					Password: "pass",
				},
				Body: domain.ComposeBody{Mode: domain.ComposeBodyMode("")},
			},
		},
		{
			name: "with raw body",
			dto: composeTemplateDTO{
				ID:     "test-6",
				Name:   "With Body",
				Method: "POST",
				URL:    "https://example.com/api",
				Body: composeBodyDTO{
					Mode: "raw",
					Raw:  `{"test": "data"}`,
				},
			},
			want: domain.ComposeRequestTemplate{
				ID:      "test-6",
				Name:    "With Body",
				Method:  "POST",
				URL:     "https://example.com/api",
				Headers: []domain.ComposeHeader{},
				Query:   []domain.ComposeQueryParam{},
				Body: domain.ComposeBody{
					Mode: domain.ComposeBodyMode("raw"),
					Raw:  `{"test": "data"}`,
				},
			},
		},
		{
			name: "with json body",
			dto: composeTemplateDTO{
				ID:     "test-7",
				Name:   "With JSON",
				Method: "POST",
				URL:    "https://example.com/api",
				Body: composeBodyDTO{
					Mode: "json",
					JSON: `{"key":"value","number":42}`,
				},
			},
			want: domain.ComposeRequestTemplate{
				ID:      "test-7",
				Name:    "With JSON",
				Method:  "POST",
				URL:     "https://example.com/api",
				Headers: []domain.ComposeHeader{},
				Query:   []domain.ComposeQueryParam{},
				Body: domain.ComposeBody{
					Mode: domain.ComposeBodyMode("json"),
					JSON: `{"key":"value","number":42}`,
				},
			},
		},
		{
			name: "with form data",
			dto: composeTemplateDTO{
				ID:     "test-8",
				Name:   "With Form",
				Method: "POST",
				URL:    "https://example.com/form",
				Body: composeBodyDTO{
					Mode: "form",
					Form: []composeFormFieldDTO{
						{Key: "username", Value: "test"},
						{Key: "password", Value: "secret"},
					},
				},
			},
			want: domain.ComposeRequestTemplate{
				ID:      "test-8",
				Name:    "With Form",
				Method:  "POST",
				URL:     "https://example.com/form",
				Headers: []domain.ComposeHeader{},
				Query:   []domain.ComposeQueryParam{},
				Body: domain.ComposeBody{
					Mode: domain.ComposeBodyMode("form"),
					Form: []domain.ComposeFormField{
						{Key: "username", Value: "test"},
						{Key: "password", Value: "secret"},
					},
				},
			},
		},
		{
			name: "with multipart",
			dto: composeTemplateDTO{
				ID:     "test-9",
				Name:   "With Multipart",
				Method: "POST",
				URL:    "https://example.com/upload",
				Body: composeBodyDTO{
					Mode: "multipart",
					Multipart: []composePartDTO{
						{
							Name:        "file",
							IsFile:      true,
							FileName:    "test.txt",
							ContentType: "text/plain",
							Base64:      "dGVzdCBkYXRh",
						},
						{
							Name:   "description",
							IsFile: false,
							Value:  "Test file",
						},
					},
				},
			},
			want: domain.ComposeRequestTemplate{
				ID:      "test-9",
				Name:    "With Multipart",
				Method:  "POST",
				URL:     "https://example.com/upload",
				Headers: []domain.ComposeHeader{},
				Query:   []domain.ComposeQueryParam{},
				Body: domain.ComposeBody{
					Mode: domain.ComposeBodyMode("multipart"),
					Multipart: []domain.ComposeMultipartPart{
						{
							Name:        "file",
							IsFile:      true,
							FileName:    "test.txt",
							ContentType: "text/plain",
							Base64:      "dGVzdCBkYXRh",
						},
						{
							Name:   "description",
							IsFile: false,
							Value:  "Test file",
						},
					},
				},
			},
		},
		{
			name: "method lowercase converted to uppercase",
			dto: composeTemplateDTO{
				ID:     "test-10",
				Name:   "Method Case",
				Method: "delete",
				URL:    "https://example.com/resource/1",
			},
			want: domain.ComposeRequestTemplate{
				ID:      "test-10",
				Name:    "Method Case",
				Method:  "DELETE",
				URL:     "https://example.com/resource/1",
				Headers: []domain.ComposeHeader{},
				Query:   []domain.ComposeQueryParam{},
				Body:    domain.ComposeBody{Mode: domain.ComposeBodyMode("")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dtoToDomainTemplate(tt.dto)

			// Compare fields except timestamps
			if got.ID != tt.want.ID {
				t.Errorf("ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Method != tt.want.Method {
				t.Errorf("Method = %v, want %v", got.Method, tt.want.Method)
			}
			if got.URL != tt.want.URL {
				t.Errorf("URL = %v, want %v", got.URL, tt.want.URL)
			}
			if !reflect.DeepEqual(got.Headers, tt.want.Headers) {
				t.Errorf("Headers = %v, want %v", got.Headers, tt.want.Headers)
			}
			if !reflect.DeepEqual(got.Query, tt.want.Query) {
				t.Errorf("Query = %v, want %v", got.Query, tt.want.Query)
			}
			if !reflect.DeepEqual(got.Auth, tt.want.Auth) {
				t.Errorf("Auth = %v, want %v", got.Auth, tt.want.Auth)
			}
			if !reflect.DeepEqual(got.Body, tt.want.Body) {
				t.Errorf("Body = %v, want %v", got.Body, tt.want.Body)
			}
			// Timestamps should be set
			if got.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set")
			}
			if got.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be set")
			}
		})
	}
}
