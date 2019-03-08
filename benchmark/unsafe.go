package benchmark

import (
	"encoding/hex"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
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
	if len(joinedElectedValidatorAddresses) != 20*len(electedValidatorIndexes) {
		panic("joinedElectedValidatorAddresses length is invalid")
	}

	return joinedElectedValidatorAddresses
}

func (h *harness) _unsafe_SetElectedValidators(senderPublicKey []byte, senderPrivateKey []byte, electedValidatorIndexes []int) error {
	payload := createPayloadForUnsafeTestsSetElectedValidators(electedValidatorIndexes)
	res, _, err := h.sendTransaction(OwnerOfAllSupply.PublicKey(), OwnerOfAllSupply.PrivateKey(), "_Elections", "unsafetests_setElectedValidators", payload)
	if err != nil {
		return errors.Wrapf(err, "UnsafeTests_SetElectedValidators() Result: +%v", res)
	}
	if res.ExecutionResult != codec.EXECUTION_RESULT_SUCCESS {
		return errors.Errorf("Failed to execute unsafe set elected validators contract. Result: %+v", res)
	}
	return nil
}
