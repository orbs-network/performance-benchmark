package benchmark

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

func TestStats(t *testing.T) {
	config := GetConfig()
	h := NewHarness(config)
	printStats(h, 0)
}

// This is not really a test, just a way to trigger this from outside

// To run from command line:
// 1. Run export API_ENDPOINT=http://35.177.173.249/vchains/1000/api/v1/ (or any other URL you want to send the command to)
// 2. Run: go test ./benchmark/... -run TestSetElectionValidator -v
func TestSetElectionValidator(t *testing.T) {
	config := GetConfig()
	h := NewHarness(config)
	elected := []int{0, 1, 2, 3, 4}
	t.Logf("Electing indices: %v", elected)
	require.Nil(t, h._unsafe_SetElectedValidators(OwnerOfAllSupply.PublicKey(), OwnerOfAllSupply.PrivateKey(), elected))
}

func TestPeriodicReelection(t *testing.T) {
	config := GetConfig()
	h := NewHarness(config)
	interval := 1 * time.Minute
	rand.Seed(time.Now().UnixNano())
	minSize := 4
	maxSize := len(StabilityNodeAddresses)

	t.Logf("===== TestPeriodicReelection start ===== checkInterval=%s reelectInterval=%s committeeSize Min=%d Max=%d\n",
		interval, REELECTION_INTERVAL, minSize, maxSize)

	for {
		committeeSize := minSize + rand.Intn(maxSize-minSize)
		maybeReelectCommittee(h, committeeSize, maxSize)
		time.Sleep(interval)
	}
}

// We use the go test functionality to run the performance test because it's easier to log with it.
// Could have called this from main() just the same.

func TestStability(t *testing.T) {
	t.Log("START")
	config := GetConfig()
	h := NewHarness(config)

	t.Logf("===== Test start ===== txCount=%d burst=%d interval=%v addressesCount=%d\n",
		config.numberOfTransactions, config.txBurstCount, config.intervalBetweenBurstsMillis, len(h.accountAddresses))
	//fastRate := rate.NewLimiter(1000, 50)

	errors := RunLoad(h, config)

	t.Logf("===== Test Completed =====")

	postTest(t, h, config, errors)
}

func postTest(t *testing.T, h *harness, config *E2EConfig, errors []error) {
	baseTxCount := getTransactionCount(t, h)
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
