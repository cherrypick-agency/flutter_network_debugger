package domain

import (
	"errors"
	"fmt"
)

var ErrRuleNotFound = errors.New("mapping rule not found")

type RuleNotFoundError struct {
	ID string
}

func (e RuleNotFoundError) Error() string {
	if e.ID == "" {
		return ErrRuleNotFound.Error()
	}
	return fmt.Sprintf("%s: %s", ErrRuleNotFound.Error(), e.ID)
}

func (e RuleNotFoundError) Unwrap() error { return ErrRuleNotFound }
