package benchmark

import (
	"fmt"
	orbsClient "github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/stretchr/testify/require"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func getTransactionCount(t *testing.T, h *harness) float64 {
	var m metrics

	require.True(t, Eventually(1*time.Minute, func() bool {
		m = h.getMetricsFromMainNode()
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
	//limiter := rate.NewLimiter(rate.Limit(config.txPerMin/60.0), 1)
	txBurst := 500
	intervalMillis := 60000 * time.Millisecond
	fmt.Printf("BURST=%d SLEEP=%s NTH=%d\n", txBurst, intervalMillis, config.metricsEveryNth)
	var i uint64
	for {
		for j := 0; j < txBurst; j++ {
			wg.Add(1)
			go func(idx uint64) {
				defer wg.Done()
				defer func() {
					if idx == 0 || idx%config.metricsEveryNth == 0 {
						printStats(h, idx)
						//printMetrics(h.getMetricsFromMainNode())
					}
				}()
				target, _ := orbsClient.CreateAccount()
				amount := uint64(ctrlRand.Intn(10))
				_, _, err2 := h.sendTransaction(OwnerOfAllSupply.PublicKey(), OwnerOfAllSupply.PrivateKey(), "BenchmarkToken", "transfer", uint64(amount), target.AddressAsBytes())
				if err2 != nil {
					errors = append(errors, err2)
				}
			}(i)
			i++
		}
		if i >= config.numberOfTransactions {
			break
		}
		time.Sleep(intervalMillis)

		//if err := limiter.Wait(context.Background()); err == nil {
	}
	wg.Wait()
	return errors
}

func printStats(h *harness, idx uint64) {
	for _, nodeIP := range h.config.allNodeIps {
		m := h.getMetrics(h.metricsEndpoint(nodeIP))
		printStatsFromMetrics(nodeIP, h.config, m, idx)
	}
	fmt.Println()
}

func printStatsFromMetrics(nodeIP string, cfg *E2EConfig, m metrics, idx uint64) (int, error) {
	return fmt.Printf("=STATS= %s %s %s txTotal=%d RateTxMin=%.0f H=%.0f currentTxIdx=%d PApiTxMaxMs=%.0f PApiTxP99Ms=%.0f SinceLastCommitMs=%.0f CommittedPoolTx=%.0f PendingPoolTx=%.0f TimeInPendingMax=%0.f TimeInPendingP99=%0.f StateKeys=%.0f BlockSyncCommittedBlocks=%.0f HeapAllocMb=%.0f Goroutines=%.0f\n",
		time.Now().UTC().Format(TIMESTAMP_FORMAT),
		m["Version.Commit"]["Value"].(string)[:8],
		nodeIP,
		cfg.numberOfTransactions,
		cfg.txPerMin,
		m["BlockStorage.BlockHeight"]["Value"],
		idx,
		m["PublicApi.SendTransactionProcessingTime"]["Max"],
		m["PublicApi.SendTransactionProcessingTime"]["P99"],
		m["ConsensusAlgo.LeanHelix.TimeSinceLastCommitMillis"]["Max"],
		m["TransactionPool.CommittedPool.TransactionCount"]["Value"],
		m["TransactionPool.PendingPool.TransactionCount"]["Value"],
		m["TransactionPool.PendingPool.TimeSpentInQueue"]["Max"],
		m["TransactionPool.PendingPool.TimeSpentInQueue"]["P99"],
		m["StateStoragePersistence.TotalNumberOfKeys"]["Value"],
		m["BlockSync.Processing.CommittedBlocks"]["Value"],
		m["Runtime.HeapAlloc"]["Value"],
		m["Runtime.NumGoroutine"]["Value"],
	)
}

func printMetrics(m metrics) (int, error) {
	//return fmt.Printf("H=%v V=%v\n",
	//	m["BlockStorage.BlockHeight"]["Value"],
	//	m["ConsensusAlgo.LeanHelix.CurrentElectionCount"]["Value"])
	return fmt.Printf("%v\n",
		m)
}

func TestStability(t *testing.T) {
	t.Log("START")
	config := getConfig()
	h := newHarness(config)

	baseTxCount := getTransactionCount(t, h)

	t.Logf("===== Test start ===== txCount=%d txPerMin=%.0f\n", config.numberOfTransactions, config.txPerMin)
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
