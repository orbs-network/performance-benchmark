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

const TEST_KEYS_FILENAME = "benchmark/addresses.json"

type harness struct {
	config           *E2EConfig
	clients          []*orbsClient.OrbsClient
	accountAddresses [][]byte
	reelectionDue    bool
	nextReelection   time.Time
}

func NewHarness(config *E2EConfig) *harness {

	clients := make([]*orbsClient.OrbsClient, 0)
	for i := 0; i < len(config.netConfig.ValidatorNodes); i++ {
		if !config.netConfig.ValidatorNodes[i].Active || len(config.netConfig.ValidatorNodes[i].IP) == 0 {
			continue
		}
		client := orbsClient.NewClient(baseUrlEndpoint(config.netConfig.ValidatorNodes[i].IP, config.vchainId), config.vchainId, codec.NETWORK_TYPE_TEST_NET)
		clients = append(clients, client)
		fmt.Printf("Added client node: %s %s\n",
			config.netConfig.ValidatorNodes[i].IP, config.netConfig.ValidatorNodes[i].Address)
	}
	fmt.Printf("Will send transactions to %d nodes\n", len(clients))

	addresses := readAddressesFromFile(TEST_KEYS_FILENAME)
	if addresses == nil {
		fmt.Printf("Addresses not loaded, perhaps %s is not found", TEST_KEYS_FILENAME)
		return nil
	}

	return &harness{
		config:           config,
		clients:          clients,
		accountAddresses: addresses,
		reelectionDue:    false,
		nextReelection:   time.Now(),
	}
}

func readAddressesFromFile(filename string) [][]byte {
	keys := getTestKeysFromFile(filename)
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

func (h *harness) deployNativeContract(from *keys.Ed25519KeyPair, contractName string, code []byte) (codec.ExecutionResult, codec.TransactionStatus, error) {
	timeoutDuration := 10 * time.Second
	beginTime := time.Now()
	sendTxOut, txId, err := h.sendTransaction(0, from.PublicKey(), from.PrivateKey(), "_Deployments", "deployService", contractName, uint32(PROCESSOR_TYPE_NATIVE), code)
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

		txStatusOut, _ := h.getTransactionStatus(0, txId)

		txStatus, executionResult = txStatusOut.TransactionStatus, txStatusOut.ExecutionResult
	}

	return executionResult, txStatus, err
}

// This sends to vchain as defined in VCHAIN environment variable
func (h *harness) sendTransaction(clientIdx int, senderPublicKey []byte, senderPrivateKey []byte, contractName string, methodName string, args ...interface{}) (response *codec.SendTransactionResponse, txId string, err error) {
	client := h.clients[clientIdx]
	payload, txId, err := client.CreateTransaction(senderPublicKey, senderPrivateKey, contractName, methodName, args...)
	if err != nil {
		return nil, txId, err
	}

	var sendTransaction func(rawTransaction []byte) (response *codec.SendTransactionResponse, err error)
	if h.config.isAsync {
		sendTransaction = client.SendTransactionAsync
	} else {
		sendTransaction = client.SendTransaction
	}
	response, err = sendTransaction(payload)
	return response, txId, err
}

func (h *harness) runQuery(clientIdx int, senderPublicKey []byte, contractName string, methodName string, args ...interface{}) (response *codec.RunQueryResponse, err error) {
	client := h.clients[clientIdx]
	payload, err := client.CreateQuery(senderPublicKey, contractName, methodName, args...)
	if err != nil {
		return nil, err
	}
	response, err = client.SendQuery(payload)
	return
}

func (h *harness) getTransactionStatus(clientIdx int, txId string) (response *codec.GetTransactionStatusResponse, err error) {
	client := h.clients[clientIdx]
	response, err = client.GetTransactionStatus(txId)
	return
}

func (h *harness) getTransactionReceiptProof(clientIdx int, txId string) (response *codec.GetTransactionReceiptProofResponse, err error) {
	client := h.clients[clientIdx]
	response, err = client.GetTransactionReceiptProof(txId)
	return
}

func baseUrlEndpoint(ip string, vchain uint32) string {
	return fmt.Sprintf("http://%s/vchains/%d", ip, vchain)
}

func metricsEndpoint(ip string, vchain uint32) string {
	return fmt.Sprintf("http://%s/vchains/%d/metrics", ip, vchain)
}

type metrics map[string]map[string]interface{}

func (h *harness) getMetricsFromMainNode() metrics {
	return h.getMetrics(metricsEndpoint(h.config.gatewayIP, h.config.vchainId))
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
