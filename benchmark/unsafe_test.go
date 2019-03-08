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

func TestCalcElected(t *testing.T) {
	for i := 0; i < 100; i++ {
		elected := calcElected(4, 7)
		t.Logf("Elected: %v", elected)
		require.Equal(t, 4, len(elected))
	}
}
