package benchmark

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

func TestCreatePayloadForUnsafeTestsSetElectedValidators(t *testing.T) {
	electedNodes := []int{0, 1}

	payload := createPayloadForUnsafeTestsSetElectedValidators(electedNodes)
	t.Logf("Payload: %s", hex.EncodeToString(payload))
	require.Equal(t, 40, len(payload))

}

func TestCalcElected(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		size := 4 + rand.Intn(3)
		elected := calcElected(size, 7)
		t.Logf("Elected: %v", elected)
		require.Equal(t, size, len(elected))
	}
}
