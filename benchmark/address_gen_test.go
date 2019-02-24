package benchmark

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// Enough to run it once to generate the addresses.json
func TestGenerateKeys(t *testing.T) {
	count := commandGenerateTestKeys()
	keys := getTestKeysFromFile()

	require.Equal(t, count, len(keys), "should have written and read same number of keys")
}
