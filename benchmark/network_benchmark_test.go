package benchmark

import (
	"context"
	"fmt"
	orbsClient "github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func getTransactionCount(t *testing.T, h *harness) float64 {
	var m metrics

	require.True(t, Eventually(1*time.Minute, func() bool {
		m = h.getMetrics()
		return m != nil
	}), "could not retrieve metrics")

	return m["TransactionPool.CommittedPool.TransactionCount"]["Value"].(float64)
}

func groupErrors(errors []error) map[string]int {
	groupedErrors := make(map[string]int)

	for _, error := range errors {
		groupedErrors[error.Error()]++
	}

	return groupedErrors
}

func testLoop(h *harness, txCount uint64, txRate *rate.Limiter) []error {
	ctrlRand := rand.New(rand.NewSource(0))
	var errors []error
	var wg sync.WaitGroup
	for i := uint64(0); i < txCount; i++ {
		if err := txRate.Wait(context.Background()); err == nil {
			wg.Add(1)
			go func(idx uint64) {
				defer wg.Done()
				defer func() {
					if idx%10 == 0 {
						fmt.Printf("%s\n", h.getMetrics()["BlockStorage.BlockHeight"])
					}
				}()
				target, _ := orbsClient.CreateAccount()
				amount := uint64(ctrlRand.Intn(10))
				_, _, err2 := h.sendTransaction(OwnerOfAllSupply.PublicKey(), OwnerOfAllSupply.PrivateKey(), "BenchmarkToken", "transfer", uint64(amount), target.AddressAsBytes())
				if err2 != nil {
					errors = append(errors, err2)
				}
			}(i)
		} else {
			errors = append(errors, err)
		}
	}
	wg.Wait()
	return errors
}

func TestE2EStress(t *testing.T) {
	config := getConfig()
	h := newHarness(config.vchainId)

	baseTxCount := getTransactionCount(t, h)

	fmt.Printf("===== Test start ===== txCount=%d", config.numberOfTransactions)
	//fastRate := rate.NewLimiter(1000, 50)
	onePerSec := rate.NewLimiter(1, 1)
	errors := testLoop(h, config.numberOfTransactions, onePerSec)

	txCount := getTransactionCount(t, h) - baseTxCount

	expectedNumberOfTx := float64(100-config.acceptableFailureRate) / 100 * float64(config.numberOfTransactions)

	fmt.Printf("Successfully processed %.0f%% of %d transactions\n", txCount/float64(config.numberOfTransactions)*100, config.numberOfTransactions)

	if len(errors) != 0 {
		fmt.Println()
		fmt.Println("===== ERRORS =====")
		for k, v := range groupErrors(errors) {
			fmt.Printf("%d times: %s\n", v, k)
		}
		fmt.Println("===== ERRORS =====")
		fmt.Println()
	}

	require.Condition(t, func() (success bool) {
		return txCount >= expectedNumberOfTx
	}, "transaction processed (%.0f) < expected transactions processed (%.0f) out of %d transactions sent", txCount, expectedNumberOfTx, config.numberOfTransactions)

	// Commenting out until we get reliable rates

	//ratePerSecond := m["TransactionPool.RatePerSecond"]["Rate"].(float64)

	//require.Condition(t, func() (success bool) {
	//	return ratePerSecond >= config.targetTPS
	//}, "actual tps (%f) is less than target tps (%f)", ratePerSecond, config.targetTPS)
}
