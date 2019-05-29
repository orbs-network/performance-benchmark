package benchmark

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

const PRINT_TX_TIME = false

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

	for _, err := range errors {
		groupedErrors[err.Error()]++
	}

	return groupedErrors
}

func RunLoad(h *harness, config *E2EConfig) []error {
	ctrlRand := rand.New(rand.NewSource(0))
	var errors []error
	var wg sync.WaitGroup
	clientsCount := len(h.clients)
	txBurst := int(config.txBurstCount)
	intervalMillis := time.Duration(config.intervalBetweenBurstsMillis) * time.Millisecond
	fmt.Printf("VCHAIN=%d ASYNC=%t BURST=%d SLEEP=%s NTH=%d ADDRESSES=%d NODE_COUNT=%d FIRST_NODE=%s OWNER_PK=%s\n",
		config.vchainId,
		config.isAsync,
		txBurst,
		intervalMillis,
		config.metricsEveryNth,
		len(h.accountAddresses),
		clientsCount,
		h.clients[0].Endpoint,
		OwnerOfAllSupply.PrivateKeyHex())
	if len(h.accountAddresses) == 0 {
		panic("No addresses were read from file")
	}
	var i uint64
	if PRINT_TX_TIME {
		fmt.Println("BLOCK_HEIGHT,BLOCK_TIMESTAMP,CLIENT_IP,RESPONSE_TIME_MILLIS")
		//fmt.Println("START_TIME,END_TIME,CLIENT_INDEX,BLOCK_HEIGHT,RESPONSE_TIME_MILLIS")
	}
	for {
		now := time.Now()
		now.Minute()
		//fmt.Printf("== %s ===> ", now.Format("15:04:05"), )
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
				target := h.accountAddresses[ctrlRand.Intn(len(h.accountAddresses))]
				//fmt.Printf("Transfer %d to address_idx #%d=%s\n", amount, addrIndex, encoding.EncodeHex(target))
				//clientIdx := 1
				clientIdx := ctrlRand.Intn(clientsCount)
				//fmt.Printf("Calling client idx=%d\n", clientIdx)
				startTime := time.Now()
				//fmt.Printf(".")
				response, _, err2 := h.sendTransaction(clientIdx, OwnerOfAllSupply.PublicKey(), OwnerOfAllSupply.PrivateKey(), "ResettableBenchmarkToken", "transfer", uint64(amount), target)
				endTime := time.Now()
				addr := h.clients[clientIdx].Endpoint
				if PRINT_TX_TIME {

					fmt.Printf("%d,%s,%.0d\n",
						//startTime.Format("15:04:05"),
						//endTime.Format("15:04:05"),
						response.BlockHeight,
						response.BlockTimestamp.Format("15:04:05"),
						//addr,
						endTime.Sub(startTime).Nanoseconds()/1000000,
						//response.TransactionStatus,
						//response.ExecutionResult,
					)
				}
				if err2 != nil {
					fmt.Printf("Error sending tx to %s: %s\n", addr, err2)
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

func maybeReelectCommittee(h *harness, committeeSize int, maxSize int) {
	now := time.Now()
	fmt.Printf("=%s Reelect?=\n", now.UTC().Format(TIMESTAMP_FORMAT))
	if now.After(h.nextReelection) {
		h.nextReelection = now.Add(REELECTION_INTERVAL)
		elected := calcElected(committeeSize, len(StabilityNodeAddresses))
		fmt.Printf("== %s Reelecting committee indices %v %d out of %d on vchain %d. Next reelection on %s\n",
			now.UTC().Format(TIMESTAMP_FORMAT),
			len(elected),
			elected,
			maxSize,
			h.clients[0].VirtualChainId,
			h.nextReelection.UTC().Format(TIMESTAMP_FORMAT))
		if len(elected) < 4 {
			fmt.Println("MUST NOT SEND LESS THAN 4 !!! SKIPPING")
			return
		}
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

func printStatsFromMetrics(nodeIP string, cfg *E2EConfig, m metrics, idx uint64) {
	var version string
	if m["Version.Commit"]["Value"] == nil {
		version = "NA"
	} else {
		version = m["Version.Commit"]["Value"].(string)[:8]
	}
	fmt.Printf("=STATS= %s PID=%d Ver=%s IP=%s RateTxMin=%d txIdx=%d Node=%s H=%.0f StateKeys=%.0f HeapAllocMb=%.0f Goroutines=%.0f\n",
		time.Now().UTC().Format(TIMESTAMP_FORMAT),
		os.Getpid(),
		version,
		nodeIP,
		cfg.txBurstCount/(cfg.intervalBetweenBurstsMillis/1000),
		idx,
		m["Node.Address"]["Value"],
		m["BlockStorage.BlockHeight"]["Value"],
		m["StateStoragePersistence.TotalNumberOfKeys.Count"]["Value"],
		m["Runtime.HeapAlloc.Bytes"]["Value"],
		m["Runtime.NumGoroutine.Number"]["Value"],
	)
}
