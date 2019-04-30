package benchmark

import (
	"github.com/stretchr/testify/require"
	"testing"
)

const TEST_KEYS_FILENAME = "addresses.json"

// Enough to run it once to generate the addresses.json
func TestGenerateKeys(t *testing.T) {
	t.Skip() // don't recreate keys by mistake
	count := commandGenerateTestKeys(TEST_KEYS_FILENAME)
	keys := getTestKeysFromFile(TEST_KEYS_FILENAME)

	require.Equal(t, count, len(keys), "should have written and read same number of keys")
}
