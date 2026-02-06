package domain

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// InterceptDirection describes interception direction
type InterceptDirection string

const (
	DirRequest  InterceptDirection = "request"
	DirResponse InterceptDirection = "response"
)

const maxRegexLength = 2048

// RuleStringMatch -- simple string field matching model
type RuleStringMatch struct {
	Equals   string   `json:"equals,omitempty"`
	Prefix   string   `json:"prefix,omitempty"`
	Suffix   string   `json:"suffix,omitempty"`
	Contains string   `json:"contains,omitempty"`
	AnyOf    []string `json:"anyOf,omitempty"`
	Regex    string   `json:"regex,omitempty"`
}

// RuleHeaderMatch -- header matching
type RuleHeaderMatch struct {
	Name  RuleStringMatch `json:"name"`
	Value RuleStringMatch `json:"value,omitempty"`
}

// RuleStatusMatch -- status ranges (for response)
type RuleStatusMatch struct {
	Equals []int `json:"equals,omitempty"`
	From   int   `json:"from,omitempty"`
	To     int   `json:"to,omitempty"`
	Is4xx  bool  `json:"is4xx,omitempty"`
	Is5xx  bool  `json:"is5xx,omitempty"`
}

// InterceptWhen -- conditions (all AND in MVP)
type InterceptWhen struct {
	Method       []string         `json:"method,omitempty"`
	Scheme       []string         `json:"scheme,omitempty"`
	Host         *RuleStringMatch `json:"host,omitempty"`
	Port         *RuleStringMatch `json:"port,omitempty"`
	Path         *RuleStringMatch `json:"path,omitempty"`
	ContentType  *RuleStringMatch `json:"contentType,omitempty"`
	Status       *RuleStatusMatch `json:"responseStatus,omitempty"`
	Header       *RuleHeaderMatch `json:"header,omitempty"`
	BodyContains string           `json:"bodyContains,omitempty"`
}

// InterceptRule -- interception rule
type InterceptRule struct {
	ID             string        `json:"id"`
	Enabled        bool          `json:"enabled"`
	Priority       int           `json:"priority"`
	Action         string        `json:"action"` // request|response|both
	Once           bool          `json:"once"`
	StopProcessing bool          `json:"stopProcessing"`
	When           InterceptWhen `json:"when"`
	CreatedAt      time.Time     `json:"createdAt,omitempty"`
	UpdatedAt      time.Time     `json:"updatedAt,omitempty"`
}

// SetDefaults sets default values for unset fields
func (r *InterceptRule) SetDefaults() {
	if r.Priority <= 0 {
		r.Priority = 10
	}
}

// Validate performs domain-level validation
func (r *InterceptRule) Validate() error {
	if r.Action == "" {
		return errors.New("action is required")
	}
	if r.Action != "request" && r.Action != "response" && r.Action != "both" {
		return errors.New("action must be one of: request, response, both")
	}
	if err := r.When.ValidateRegex(); err != nil {
		return fmt.Errorf("when: %w", err)
	}
	return nil
}

// ValidateRegex checks that all regex patterns in the condition are valid
func (w *InterceptWhen) ValidateRegex() error {
	check := func(name string, m *RuleStringMatch) error {
		if m != nil && m.Regex != "" {
			if len(m.Regex) > maxRegexLength {
				return fmt.Errorf("%s regex too long (max %d bytes)", name, maxRegexLength)
			}
			if _, err := regexp.Compile(m.Regex); err != nil {
				return fmt.Errorf("%s regex invalid: %w", name, err)
			}
		}
		return nil
	}
	if err := check("host", w.Host); err != nil {
		return err
	}
	if err := check("port", w.Port); err != nil {
		return err
	}
	if err := check("path", w.Path); err != nil {
		return err
	}
	if err := check("contentType", w.ContentType); err != nil {
		return err
	}
	if w.Header != nil {
		if err := check("header.name", &w.Header.Name); err != nil {
			return err
		}
		if err := check("header.value", &w.Header.Value); err != nil {
			return err
		}
	}
	return nil
}
