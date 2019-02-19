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
	"testing"
	"time"
)

type harness struct {
	config *E2EConfig
	client *orbsClient.OrbsClient
}

func newHarness(config *E2EConfig) *harness {
	return &harness{
		config: config,
		client: orbsClient.NewClient(getConfig().baseUrl, config.vchainId, codec.NETWORK_TYPE_TEST_NET),
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

func (h *harness) metricsEndpoint(ip string) string {
	vchainIdStr := fmt.Sprintf("%d", h.config.vchainId)
	return "http://" + ip + "/vchains/" + vchainIdStr + "/metrics"
}

type metrics map[string]map[string]interface{}

func (h *harness) getMetricsFromMainNode() metrics {
	return h.getMetrics(h.absoluteUrlFor("/metrics"))
}

func (h *harness) getMetrics(metricsUrl string) metrics {

	res, err := http.Get(metricsUrl)

	if err != nil {
		fmt.Printf("Failed to read metrics from %s: %s\n", metricsUrl, err.Error())
		return nil
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
		m := h.getMetricsFromMainNode()
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
