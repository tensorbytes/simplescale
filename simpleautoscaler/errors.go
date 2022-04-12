package simpleautoscaler

import (
	"errors"
)

func IsNotFoundPolicyField(err error) bool {
	return errors.Is(err, ErrNotFoundPolicyField)
}
