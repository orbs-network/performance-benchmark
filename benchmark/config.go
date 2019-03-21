package benchmark

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type E2EConfig struct {
	vchainId uint32
	//baseUrl          string
	gatewayIP        string
	netConfig        *nodeConfiguration
	ethereumEndpoint string

	StressTestConfig
}

type StressTestConfig struct {
	allNodeIps                  []string
	numberOfTransactions        uint64
	acceptableFailureRate       uint64
	targetTPS                   float64
	metricsEveryNth             uint64
	txBurstCount                uint64
	intervalBetweenBurstsMillis uint64
}

const TIMESTAMP_FORMAT = "2006-01-02T15:04:05.000Z"

const VIRTUAL_CHAIN_ID = uint32(2000)

// "github.com/orbs-network/orbs-spec/types/go/protocol"
const PROCESSOR_TYPE_NATIVE = 1

const REELECTION_INTERVAL = 10 * time.Minute

func getConfig() *E2EConfig {
	vchainId := VIRTUAL_CHAIN_ID
	//baseUrl := "http://54.194.120.89:8080"
	//baseUrl := "http://localhost:8080"

	stressTestNumberOfTransactions := uint64(1000000000)
	stressTestFailureRate := uint64(20)
	stressTestTargetTPS := float64(20)
	stressTestMetricsEveryNthTransaction := uint64(100)

	stressTestTxBurstCount := uint64(1)
	stressTestIntervalBetweenBurstsMillis := uint64(300000)

	undefinedVars := make([]string, 0)
	for _, envVar := range []string{"NODE_IPS", "VCHAIN", "TX_BURST_COUNT", "INTERVAL_BETWEEN_BURSTS_MILLIS", "METRICS_EVERY_NTH_TX", "GATEWAY_IP", "NET_CONFIG_URL"} {
		if os.Getenv(envVar) == "" {
			undefinedVars = append(undefinedVars, envVar)
		}
	}

	if len(undefinedVars) > 0 {
		panic(fmt.Sprintf("Must define environment variables %s", strings.Join(undefinedVars, ",")))
	}

	gatewayIP := os.Getenv("GATEWAY_IP")
	ethereumEndpoint := os.Getenv("ETHEREUM_ENDPOINT")

	if vcid, err := strconv.ParseUint(os.Getenv("VCHAIN"), 10, 0); err == nil {
		vchainId = uint32(vcid)
	} else {
		panic("Environment variable VCHAIN must be a number")
	}

	if txBurstCount, err := strconv.ParseUint(os.Getenv("TX_BURST_COUNT"), 10, 0); err == nil {
		stressTestTxBurstCount = txBurstCount
	}

	if intervalBetweenBurstsMillis, err := strconv.ParseUint(os.Getenv("INTERVAL_BETWEEN_BURSTS_MILLIS"), 10, 0); err == nil {
		stressTestIntervalBetweenBurstsMillis = intervalBetweenBurstsMillis
	}

	if numTx, err := strconv.ParseUint(os.Getenv("STRESS_TEST_NUMBER_OF_TRANSACTIONS"), 10, 0); err == nil {
		stressTestNumberOfTransactions = numTx
	}

	if metricsEveryNth, err := strconv.ParseUint(os.Getenv("METRICS_EVERY_NTH_TX"), 10, 0); err == nil {
		stressTestMetricsEveryNthTransaction = uint64(metricsEveryNth)
	}
	if failRate, err := strconv.ParseUint(os.Getenv("STRESS_TEST_FAILURE_RATE"), 10, 0); err == nil {
		stressTestFailureRate = failRate
	}

	if tps, err := strconv.ParseFloat(os.Getenv("STRESS_TEST_TARGET_TPS"), 0); err == nil {
		stressTestTargetTPS = tps
	}

	allNodeIpsStr := os.Getenv("NODE_IPS")
	if len(allNodeIpsStr) < 4 {
		panic("Must define at least 4 nodes in NODE_IPS environment variable (comma-separated list of IPs)")
	}

	netConfig, err := ReadUrlConfig(os.Getenv("NET_CONFIG_URL"))
	if err != nil {
		panic(fmt.Sprintf("Failed parsing file NET_CONFIG_URL=%s", os.Getenv("NET_CONFIG_URL")))
	}

	allNodeIps := strings.Split(allNodeIpsStr, ",")

	return &E2EConfig{
		vchainId,
		gatewayIP,
		netConfig,
		ethereumEndpoint,

		StressTestConfig{
			allNodeIps:                  allNodeIps,
			numberOfTransactions:        stressTestNumberOfTransactions,
			acceptableFailureRate:       stressTestFailureRate,
			targetTPS:                   stressTestTargetTPS,
			metricsEveryNth:             stressTestMetricsEveryNthTransaction,
			txBurstCount:                stressTestTxBurstCount,
			intervalBetweenBurstsMillis: stressTestIntervalBetweenBurstsMillis,
		},
	}
}
