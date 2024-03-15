package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIDTypeValue(t *testing.T) {
	assert.Equal(t, "type", ID("type:val").Type())
	assert.Equal(t, "val", ID("type:val").ID())
}
