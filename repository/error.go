package repository

import (
	"errors"
	"fmt"
)

// WrapErrf and WrapErr is a thin wrapper around fmt.Errorf and errors.New for consistent error creation.
// Example: return WrapErrf("failed to fetch estate", err)
// We can extend this wrapper to use stacktrace or something to handle the error based on foundation level
func WrapErrf(msg string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// WrapErr is a wrapper around errors.New
func WrapErr(msg string) error {
	return errors.New(msg)
}
