package blockquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestNewStringResult(t *testing.T) {
	source := []byte(`
{
		"key": "value"
}`)
	results := NewStringResults("value")
	got := gjson.GetBytes(source, "key")
	assert.Equal(t, results[0].Value(), got.Value())
}

func TestNewIntResult(t *testing.T) {
	source := []byte(`
{
		"key": 1
}`)
	results := NewIntResults(1)
	got := gjson.GetBytes(source, "key")
	assert.Equal(t, results[0].Value(), got.Value())
}
