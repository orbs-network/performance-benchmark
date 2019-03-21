package benchmark

import (
	"encoding/hex"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/pkg/errors"
	"math/rand"
	"time"
)

func createPayloadForUnsafeTestsSetElectedValidators(electedValidatorIndexes []int) []byte {
	joinedElectedValidatorAddresses := []byte{}
	for _, electedValidatorIndex := range electedValidatorIndexes {
		addr := StabilityNodeAddresses[electedValidatorIndex]
		address, err := hex.DecodeString(addr)
		if err != nil {

		}
		joinedElectedValidatorAddresses = append(joinedElectedValidatorAddresses, address...)
	}
	if len(joinedElectedValidatorAddresses) != 20*len(electedValidatorIndexes) {
		panic("joinedElectedValidatorAddresses length is invalid")
	}

	return joinedElectedValidatorAddresses
}

// Make sure to send to the correct vchain
// Determined by VCHAIN environment variable
func (h *harness) _unsafe_SetElectedValidators(senderPublicKey []byte, senderPrivateKey []byte, electedValidatorIndexes []int) error {
	payload := createPayloadForUnsafeTestsSetElectedValidators(electedValidatorIndexes)
	res, _, err := h.sendTransaction(0, OwnerOfAllSupply.PublicKey(), OwnerOfAllSupply.PrivateKey(), "_Elections", "unsafetests_setElectedValidators", payload)
	if err != nil {
		return errors.Wrapf(err, "UnsafeTests_SetElectedValidators() Result: %v", res)
	}
	if res.ExecutionResult != codec.EXECUTION_RESULT_SUCCESS {
		return errors.Errorf("Failed to execute unsafe set elected validators contract. Result: %+v", res)
	}
	return nil
}

func calcElected(subset int, total int) []int {
	rand.Seed(time.Now().UnixNano())
	return rand.Perm(total)[:subset]
}

// 0 1 2 3 4 5 6
// 2 3 5 6
// 0 2 4 5
// 1 2 3 4
