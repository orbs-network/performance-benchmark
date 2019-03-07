package benchmark

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreatePayloadForUnsafeTestsSetElectedValidators(t *testing.T) {
	electedNodes := []int{0, 1}

	payload := createPayloadForUnsafeTestsSetElectedValidators(electedNodes)
	t.Logf("Payload: %s", hex.EncodeToString(payload))
	require.Equal(t, 40, len(payload))

}
