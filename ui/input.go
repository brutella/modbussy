package ui

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

type NumberAccessor[T Number] struct {
	val *T
}

func NewNumberAccessor[T Number](val *T) *NumberAccessor[T] {
	return &NumberAccessor[T]{val}
}

func (a *NumberAccessor[T]) Get() string {
	if a.val == nil {
		return ""
	}
	return fmt.Sprintf("%v", *a.val)
}

func (a *NumberAccessor[T]) Set(value string) {
	i, err := strconv.Atoi(value)
	if err == nil {
		*a.val = T(i)
	}
}

// newIntInput returns an input which validates the value
// against a min and max value.
func newIntInput(min, max int) *huh.Input {
	return huh.NewInput().Validate(func(s string) error {
		i, err := strconv.Atoi(s)
		if err != nil {
			return errors.New("input not a number")
		}

		if i < min || i > max {
			return fmt.Errorf("input must be between %d and %d", min, max)
		}

		return nil
	}).Placeholder(fmt.Sprintf("%d-%d", min, max))
}
