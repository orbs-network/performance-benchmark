package benchmark

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreatePayloadForUnsafeTestsSetElectedValidators(t *testing.T) {
	electedNodes := []int{0, 1}

	payload := createPayloadForUnsafeTestsSetElectedValidators(electedNodes)
	require.Equal(t, 40, len(payload))

}
