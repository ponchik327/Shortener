package service

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var base62Re = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func TestGenerateCode_Length(t *testing.T) {
	for _, l := range []int{6, 8, 10, 16} {
		code, err := generateCode(l)
		require.NoError(t, err)
		assert.Len(t, code, l)
	}
}

func TestGenerateCode_OnlyBase62Chars(t *testing.T) {
	for range 200 {
		code, err := generateCode(8)
		require.NoError(t, err)
		assert.Regexp(t, base62Re, code)
	}
}

func TestGenerateCode_Uniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for range 1000 {
		code, err := generateCode(8)
		require.NoError(t, err)
		_, dup := seen[code]
		assert.False(t, dup, "duplicate code: %s", code)
		seen[code] = struct{}{}
	}
}
