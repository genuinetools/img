package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// validateHasNoArgs is used for commands that should not be given arguments
func validateHasNoArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}

	return errors.New("no arguments expected")
}

func allowAnyValue(value string) (string, error) {
	return value, nil
}

// listValue holds a list of values and a validation function.
type listValue struct {
	values    []string
	validator validatorFunc
}

// newListValue creates a new listValue with the specified validator.
func newListValue() *listValue {
	var values []string
	return &listValue{
		values:    values,
		validator: allowAnyValue,
	}
}

func (opts *listValue) String() string {
	if len(opts.values) == 0 {
		return ""
	}

	return fmt.Sprintf("%v", opts.values)
}

// Set validates if needed the input value and adds it to the
// internal slice.
func (opts *listValue) Set(value string) error {
	if opts.validator != nil {
		v, err := opts.validator(value)
		if err != nil {
			return err
		}
		value = v
	}
	opts.values = append(opts.values, value)
	return nil
}

// GetAll returns the values of slice.
func (opts *listValue) GetAll() []string {
	return opts.values
}

// Len returns the amount of element in the slice.
func (opts *listValue) Len() int {
	return len(opts.values)
}

// Type returns a string name for this Option type
func (opts *listValue) Type() string {
	return "list"
}

// WithValidator returns the listValue with validator set.
func (opts *listValue) WithValidator(validator validatorFunc) *listValue {
	opts.validator = validator
	return opts
}

// validatorFunc defines a validator function that returns a validated string and/or an error.
type validatorFunc func(val string) (string, error)
