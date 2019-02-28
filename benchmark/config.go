package benchmark

import (
	"os"
	"strconv"
	"strings"
)

type E2EConfig struct {
	vchainId         uint32
	baseUrl          string
	ethereumEndpoint string

	StressTestConfig
}

type StressTestConfig struct {
	allNodeIps            []string
	numberOfTransactions  uint64
	acceptableFailureRate uint64
	targetTPS             float64
	txPerMin              float64
	metricsEveryNth       uint64
}

const TIMESTAMP_FORMAT = "2006-01-02T15:04:05.000Z"

const VIRTUAL_CHAIN_ID = uint32(2000)

// "github.com/orbs-network/orbs-spec/types/go/protocol"
const PROCESSOR_TYPE_NATIVE = 1

func getConfig() *E2EConfig {
	vchainId := VIRTUAL_CHAIN_ID
	//baseUrl := "http://54.194.120.89:8080"
	baseUrl := "http://localhost:8080"

	stressTestNumberOfTransactions := uint64(1000000000)
	stressTestTransactionsPerMinute := float64(240)
	stressTestFailureRate := uint64(20)
	stressTestTargetTPS := float64(20)
	stressTestMetricsEveryNthTransaction := uint64(100)
	stressTestAllNodeIps := "54.194.120.89 35.177.173.249 52.47.211.186 35.174.231.96 18.191.62.179 52.60.152.22 18.195.172.240"

	ethereumEndpoint := "http://127.0.0.1:8545"

	if os.Getenv("API_ENDPOINT") != "" {
		apiEndpoint := os.Getenv("API_ENDPOINT")
		baseUrl = strings.TrimRight(strings.TrimRight(apiEndpoint, "/"), "/api/v1")
		ethereumEndpoint = os.Getenv("ETHEREUM_ENDPOINT")
	}

	if vcid, err := strconv.ParseUint(os.Getenv("VCHAIN"), 10, 0); err == nil {
		vchainId = uint32(vcid)
	}

	if numTx, err := strconv.ParseUint(os.Getenv("STRESS_TEST_NUMBER_OF_TRANSACTIONS"), 10, 0); err == nil {
		stressTestNumberOfTransactions = numTx
	}

	if txPerMin, err := strconv.ParseUint(os.Getenv("STRESS_TEST_TRANSACTIONS_PER_MINUTE"), 10, 0); err == nil {
		stressTestTransactionsPerMinute = float64(txPerMin)
	}

	if metricsEveryNth, err := strconv.ParseUint(os.Getenv("STRESS_TEST_METRICS_EVERY_NTH_TRANSACTION"), 10, 0); err == nil {
		stressTestMetricsEveryNthTransaction = uint64(metricsEveryNth)
	}
	if failRate, err := strconv.ParseUint(os.Getenv("STRESS_TEST_FAILURE_RATE"), 10, 0); err == nil {
		stressTestFailureRate = failRate
	}

	if tps, err := strconv.ParseFloat(os.Getenv("STRESS_TEST_TARGET_TPS"), 0); err == nil {
		stressTestTargetTPS = tps
	}

	allNodeIpsStr := os.Getenv("STRESS_TEST_ALL_NODE_IPS")
	if len(allNodeIpsStr) > 0 {
		stressTestAllNodeIps = allNodeIpsStr
	}
	allNodeIps := strings.Split(stressTestAllNodeIps, " ")

	return &E2EConfig{
		vchainId,
		baseUrl,
		ethereumEndpoint,

		StressTestConfig{
			allNodeIps:            allNodeIps,
			numberOfTransactions:  stressTestNumberOfTransactions,
			acceptableFailureRate: stressTestFailureRate,
			targetTPS:             stressTestTargetTPS,
			txPerMin:              stressTestTransactionsPerMinute,
			metricsEveryNth:       stressTestMetricsEveryNthTransaction,
		},
	}
}
