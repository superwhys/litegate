package proxy

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestParsePlace(t *testing.T) {
	place, name := parsePlace("$header.token")
	t.Logf("place: %s, name: %s", place, name)
	assert.Equal(t, "$header", place)
	assert.Equal(t, "token", name)

	place, name = parsePlace("$query.token")
	t.Logf("place: %s, name: %s", place, name)
	assert.Equal(t, "$query", place)
	assert.Equal(t, "token", name)

	place, name = parsePlace("token")
	t.Logf("place: %s, name: %s", place, name)
	assert.Equal(t, "", place)
	assert.Equal(t, "token", name)
}
