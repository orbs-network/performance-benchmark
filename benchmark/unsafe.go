package benchmark

import (
	"encoding/hex"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-network-go/crypto/digest"
	"github.com/pkg/errors"
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
	if len(joinedElectedValidatorAddresses) != digest.NODE_ADDRESS_SIZE_BYTES*len(electedValidatorIndexes) {
		panic("joinedElectedValidatorAddresses length is invalid")
	}

	return joinedElectedValidatorAddresses
}

func (h *harness) _unsafe_SetElectedValidators(senderPublicKey []byte, senderPrivateKey []byte, electedValidatorIndexes []int) error {
	payload := createPayloadForUnsafeTestsSetElectedValidators(electedValidatorIndexes)
	res, _, err := h.sendTransaction(OwnerOfAllSupply.PublicKey(), OwnerOfAllSupply.PrivateKey(), "_Elections", "unsafetests_setElectedValidators", payload)
	if err != nil {
		return errors.Wrap(err, "UnsafeTests_SetElectedValidators()")
	}
	if res.ExecutionResult != codec.EXECUTION_RESULT_SUCCESS {
		return errors.Errorf("Failed to execute unsafe set elected validators contract. Result: %s", res.ExecutionResult)
	}
	return nil
}
