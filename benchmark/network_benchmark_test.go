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

type testConfig struct {
	txCount         uint64
	txPerSec        float64
	metricsEveryNth uint64
}

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

func runTest(h *harness, config *E2EConfig) []error {
	ctrlRand := rand.New(rand.NewSource(0))
	var errors []error
	var wg sync.WaitGroup
	limiter := rate.NewLimiter(rate.Limit(config.txPerMin/60.0), 1)
	for i := uint64(0); i < config.numberOfTransactions; i++ {
		if err := limiter.Wait(context.Background()); err == nil {
			wg.Add(1)
			go func(idx uint64) {
				defer wg.Done()
				defer func() {
					if idx%config.metricsEveryNth == 0 {
						printMetrics(h.getMetrics())
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

func printMetrics(m metrics) (int, error) {
	//return fmt.Printf("H=%v V=%v\n",
	//	m["BlockStorage.BlockHeight"]["Value"],
	//	m["ConsensusAlgo.LeanHelix.CurrentElectionCount"]["Value"])
	return fmt.Printf("%v\n",
		m)
}

func TestStability(t *testing.T) {
	config := getConfig()
	h := newHarness(config.vchainId)

	baseTxCount := getTransactionCount(t, h)

	fmt.Printf("===== Test start ===== txCount=%d txPerMin=%f\n", config.numberOfTransactions, config.txPerMin)
	//fastRate := rate.NewLimiter(1000, 50)

	errors := runTest(h, config)
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
