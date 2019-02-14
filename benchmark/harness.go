package benchmark

import (
	"encoding/json"
	"fmt"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/crypto/keys"
	orbsClient "github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type E2EConfig struct {
	vchainId         uint32
	baseUrl          string
	ethereumEndpoint string

	StressTestConfig
}

type StressTestConfig struct {
	numberOfTransactions  uint64
	acceptableFailureRate uint64
	targetTPS             float64
	txPerMin              float64
	metricsEveryNth       uint64
}

const VIRTUAL_CHAIN_ID = uint32(42)

// "github.com/orbs-network/orbs-spec/types/go/protocol"
const PROCESSOR_TYPE_NATIVE = 1

type harness struct {
	client *orbsClient.OrbsClient
}

func newHarness(vchainId uint32) *harness {
	return &harness{
		client: orbsClient.NewClient(getConfig().baseUrl, vchainId, codec.NETWORK_TYPE_TEST_NET),
	}
}

func (h *harness) deployNativeContract(from *keys.Ed25519KeyPair, contractName string, code []byte) (codec.ExecutionResult, codec.TransactionStatus, error) {
	timeoutDuration := 10 * time.Second
	beginTime := time.Now()
	sendTxOut, txId, err := h.sendTransaction(from.PublicKey(), from.PrivateKey(), "_Deployments", "deployService", contractName, uint32(PROCESSOR_TYPE_NATIVE), code)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to deploy native contract")
	}

	txStatus, executionResult := sendTxOut.TransactionStatus, sendTxOut.ExecutionResult

	for txStatus == codec.TRANSACTION_STATUS_PENDING {
		// check timeout
		if time.Now().Sub(beginTime) > timeoutDuration {
			return "", "", fmt.Errorf("contract deployment is TRANSACTION_STATUS_PENDING for over %v", timeoutDuration)
		}

		time.Sleep(10 * time.Millisecond)

		txStatusOut, _ := h.getTransactionStatus(txId)

		txStatus, executionResult = txStatusOut.TransactionStatus, txStatusOut.ExecutionResult
	}

	return executionResult, txStatus, err
}

func (h *harness) sendTransaction(senderPublicKey []byte, senderPrivateKey []byte, contractName string, methodName string, args ...interface{}) (response *codec.SendTransactionResponse, txId string, err error) {
	payload, txId, err := h.client.CreateTransaction(senderPublicKey, senderPrivateKey, contractName, methodName, args...)
	if err != nil {
		return nil, txId, err
	}
	response, err = h.client.SendTransaction(payload)
	return
}

func (h *harness) runQuery(senderPublicKey []byte, contractName string, methodName string, args ...interface{}) (response *codec.RunQueryResponse, err error) {
	payload, err := h.client.CreateQuery(senderPublicKey, contractName, methodName, args...)
	if err != nil {
		return nil, err
	}
	response, err = h.client.SendQuery(payload)
	return
}

func (h *harness) getTransactionStatus(txId string) (response *codec.GetTransactionStatusResponse, err error) {
	response, err = h.client.GetTransactionStatus(txId)
	return
}

func (h *harness) getTransactionReceiptProof(txId string) (response *codec.GetTransactionReceiptProofResponse, err error) {
	response, err = h.client.GetTransactionReceiptProof(txId)
	return
}

func (h *harness) absoluteUrlFor(endpoint string) string {
	return getConfig().baseUrl + endpoint
}

type metrics map[string]map[string]interface{}

func (h *harness) getMetrics() metrics {
	res, err := http.Get(h.absoluteUrlFor("/metrics"))

	if err != nil {
		fmt.Println(h.absoluteUrlFor("/metrics"), err)
	}

	if res == nil {
		return nil
	}

	readBytes, _ := ioutil.ReadAll(res.Body)
	m := make(metrics)
	json.Unmarshal(readBytes, &m)

	return m
}

func (h *harness) waitUntilTransactionPoolIsReady(t *testing.T) {
	require.True(t, Eventually(3*time.Second, func() bool { // 3 seconds to avoid jitter but it really shouldn't take that long
		m := h.getMetrics()
		if m == nil {
			return false
		}

		blockHeight := m["TransactionPool.BlockHeight"]["Value"].(float64)

		return blockHeight > 0
	}), "Timed out waiting for metric TransactionPool.BlockHeight > 0")
}

func printTestTime(t *testing.T, msg string, last *time.Time) {
	t.Logf("%s (+%.3fs)", msg, time.Since(*last).Seconds())
	*last = time.Now()
}

func getConfig() *E2EConfig {
	vchainId := VIRTUAL_CHAIN_ID
	baseUrl := "http://localhost:8080"

	stressTestNumberOfTransactions := uint64(1000)
	stressTestTransactionsPerMinute := float64(240)
	stressTestFailureRate := uint64(20)
	stressTestTargetTPS := float64(20)
	stressTestMetricsEveryNthTransaction := uint64(10)

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

	return &E2EConfig{
		vchainId,
		baseUrl,
		ethereumEndpoint,

		StressTestConfig{
			numberOfTransactions:  stressTestNumberOfTransactions,
			acceptableFailureRate: stressTestFailureRate,
			targetTPS:             stressTestTargetTPS,
			txPerMin:              stressTestTransactionsPerMinute,
			metricsEveryNth:       stressTestMetricsEveryNthTransaction,
		},
	}
}
