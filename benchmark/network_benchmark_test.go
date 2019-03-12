package benchmark

import (
	"fmt"
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

	if m["TransactionPool.CommittedPool.Transactions.Count"]["Value"] == nil {
		return 0
	}
	return m["TransactionPool.CommittedPool.Transactions.Count"]["Value"].(float64)
}

func groupErrors(errors []error) map[string]int {
	groupedErrors := make(map[string]int)

	for _, error := range errors {
		groupedErrors[error.Error()]++
	}

	return groupedErrors
}

func runTest(h *harness, config *E2EConfig, addresses [][]byte) []error {
	ctrlRand := rand.New(rand.NewSource(0))
	var errors []error
	var wg sync.WaitGroup
	//limiter := rate.NewLimiter(rate.Limit(config.txPerMin/60.0), 1)
	txBurst := int(config.txBurstCount)
	intervalMillis := time.Duration(config.intervalBetweenBurstsMillis) * time.Millisecond
	fmt.Printf("BURST=%d SLEEP=%s NTH=%d ADDRESSES=%d TO_URL=%s OWNER_PK=%s\n",
		txBurst,
		intervalMillis,
		config.metricsEveryNth,
		len(addresses),
		h.client.Endpoint,
		OwnerOfAllSupply.PrivateKeyHex())
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
				amount := uint64(ctrlRand.Intn(5))
				addrIndex := ctrlRand.Intn(len(addresses))
				target := addresses[addrIndex]
				//fmt.Printf("Transfer %d to address_idx #%d=%s\n", amount, addrIndex, encoding.EncodeHex(target))
				_, _, err2 := h.sendTransaction(OwnerOfAllSupply.PublicKey(), OwnerOfAllSupply.PrivateKey(), "BenchmarkToken", "transfer", uint64(amount), target)
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

func maybeReelectCommittee(h *harness, committeeSize int) {
	now := time.Now()
	fmt.Printf("=%s Reelect?=\n", now.UTC().Format(TIMESTAMP_FORMAT))
	if now.After(h.nextReelection) {
		h.nextReelection = now.Add(REELECTION_INTERVAL)
		elected := calcElected(committeeSize, len(StabilityNodeAddresses))
		fmt.Printf("== %s Reelecting committee indices %v on vchain %d. Next reelection on %s\n",
			now.UTC().Format(TIMESTAMP_FORMAT), elected, h.client.VirtualChainId, h.nextReelection.UTC().Format(TIMESTAMP_FORMAT))
		err := h._unsafe_SetElectedValidators(OwnerOfAllSupply.PublicKey(), OwnerOfAllSupply.PrivateKey(), elected)
		if err != nil {
			fmt.Printf("Error electing %v, next in %s", elected, h.nextReelection)
		} else {
			fmt.Printf("== %s Success electing committee %v, next in %s\n", now.UTC().Format(TIMESTAMP_FORMAT), elected, h.nextReelection.UTC().Format(TIMESTAMP_FORMAT))
		}
	}
}

func printStats(h *harness, idx uint64) {
	for _, nodeIP := range h.config.allNodeIps {
		m := h.getMetrics(metricsEndpoint(nodeIP, h.config.vchainId))
		printStatsFromMetrics(nodeIP, h.config, m, idx)
	}
	fmt.Println()
}

func printStatsFromMetrics(nodeIP string, cfg *E2EConfig, m metrics, idx uint64) (int, error) {
	var version string
	if m["Version.Commit"]["Value"] == nil {
		version = "NA"
	} else {
		version = m["Version.Commit"]["Value"].(string)[:8]
	}
	return fmt.Printf("=STATS= %s Ver=%s IP=%s txTotal=%d RateTxMin=%d currentTxIdx=%d Node=%s H=%.0f PApiTxMaxMs=%.0f PApiTxP99Ms=%.0f SinceLastCommitMs=%.0f CommittedPoolTx=%.0f PendingPoolTx=%.0f TimeInPendingMax=%0.f TimeInPendingP99=%0.f StateKeys=%.0f BlockSyncCommittedBlocks=%.0f HeapAllocMb=%.0f Goroutines=%.0f\n",
		time.Now().UTC().Format(TIMESTAMP_FORMAT),
		version,
		nodeIP,
		cfg.numberOfTransactions,
		cfg.txBurstCount/(cfg.intervalBetweenBurstsMillis/1000),
		idx,
		m["Node.Address"]["Value"],
		m["BlockStorage.BlockHeight"]["Value"],
		m["PublicApi.SendTransactionProcessingTime.Millis"]["Max"],
		m["PublicApi.SendTransactionProcessingTime.Millis"]["P99"],
		m["ConsensusAlgo.LeanHelix.TimeSinceLastCommit.Millis"]["Max"],
		m["TransactionPool.CommittedPool.Transactions.Count"]["Value"],
		m["TransactionPool.PendingPool.Transactions.Count"]["Value"],
		m["TransactionPool.PendingPool.TimeSpentInQueue.Millis"]["Max"],
		m["TransactionPool.PendingPool.TimeSpentInQueue.Millis"]["P99"],
		m["StateStoragePersistence.TotalNumberOfKeys.Count"]["Value"],
		m["BlockSync.ProcessingBlocksState.CommittedBlocks.Count"]["Value"],
		m["Runtime.HeapAlloc.Bytes"]["Value"],
		m["Runtime.NumGoroutine.Value"]["Value"],
	)
}

func TestStats(t *testing.T) {
	config := getConfig()
	h := newHarness(config)
	printStats(h, 0)
}

// This is not really a test, just a way to trigger this from outside

// To run from command line:
// 1. Run export API_ENDPOINT=http://35.177.173.249/vchains/1000/api/v1/ (or any other URL you want to send the command to)
// 2. Run: go test ./benchmark/... -run TestSetElectionValidator -v
func TestSetElectionValidator(t *testing.T) {
	config := getConfig()
	h := newHarness(config)
	elected := []int{0, 1, 2, 3, 4}
	t.Logf("Electing indices: %v", elected)
	require.Nil(t, h._unsafe_SetElectedValidators(OwnerOfAllSupply.PublicKey(), OwnerOfAllSupply.PrivateKey(), elected))
}

func TestPeriodicReelection(t *testing.T) {
	config := getConfig()
	h := newHarness(config)
	interval := 1 * time.Minute
	rand.Seed(time.Now().UnixNano())
	minSize := 4
	maxSize := 7

	t.Logf("===== TestPeriodicReelection start ===== checkInterval=%s reelectInterval=%s committeeSize Min=%d Max=%d\n",
		interval, REELECTION_INTERVAL, minSize, maxSize)

	for {
		committeeSize := minSize + rand.Intn(maxSize-minSize)
		maybeReelectCommittee(h, committeeSize)
		time.Sleep(interval)
	}
}

func TestStability(t *testing.T) {
	t.Log("START")
	config := getConfig()
	h := newHarness(config)

	addresses := readAddressesFromFile()
	if addresses == nil {
		t.Errorf("Addresses not loaded, perhaps %s is not found", TEST_KEYS_FILENAME)
		return
	}

	baseTxCount := getTransactionCount(t, h)

	t.Logf("===== Test start ===== txCount=%d txPerMin=%.0f addressesCount=%d\n",
		config.numberOfTransactions, config.txPerMin, len(addresses))
	//fastRate := rate.NewLimiter(1000, 50)

	errors := runTest(h, config, addresses)

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

func readAddressesFromFile() [][]byte {
	keys := getTestKeysFromFile()
	if keys == nil {
		return nil
	}

	addresses := make([][]byte, 0)
	for _, key := range keys {
		addresses = append(addresses, key.Address)
		//fmt.Printf("Added to addresses array: %s\n", encoding.EncodeHex(key.Address))
	}
	return addresses
}
