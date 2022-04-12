package simpleautoscaler

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrors(t *testing.T) {
	err1 := fmt.Errorf("test error wapper error. %w", ErrNotFoundPolicyField)
	err2 := fmt.Errorf("new error wapper for it. %w", err1)
	assert.Equal(t, IsNotFoundPolicyField(err1), true, "check error wapper")
	assert.Equal(t, IsNotFoundPolicyField(err2), true, "check error wapper")
}
