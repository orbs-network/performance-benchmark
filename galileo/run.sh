#!/bin/bash

# Config for stability network
export VCHAIN=${VCHAIN-2001}
export GATEWAY_IP=${GATEWAY_IP-35.177.173.249}
export TX_BURST_COUNT=${TX_BURST_COUNT-500}
export INTERVAL_BETWEEN_BURSTS_MILLIS=${INTERVAL_BETWEEN_BURSTS_MILLIS-60000}
export METRICS_EVERY_NTH_TX=${METRICS_EVERY_NTH_TX-5000}
export ETHEREUM_ENDPOINT=${ETHEREUM_ENDPOINT-http://127.0.0.1:8545}
export NET_CONFIG_URL=${NET_CONFIG_URL-https://s3.eu-central-1.amazonaws.com/boyar-stability/boyar/config.json}

# Config for validators network
# export VCHAIN=2019
# export API_ENDPOINT=${API_ENDPOINT-http://3.209.8.117/vchains/${VCHAIN}/api/v1/}
#export TX_BURST_COUNT=1
#export INTERVAL_BETWEEN_BURSTS_MILLIS=300000
#export STRESS_TEST_METRICS_EVERY_NTH_TRANSACTION=1

# This is the IP we are sending transactions too, make sure it stays up!
#export BASE_URL=${BASE_URL-http://35.177.173.249/vchains/1000}
export COMMIT=${COMMIT-master}
DATE=$(date +%Y-%m-%d-%H%M%S)
export RESULTS=${RESULTS-results/$COMMIT/$DATE}

# !!This is used only for collecting metrics, it does not affect which IPs are on the network (every node has a static list of its peers)!!
#export STRESS_TEST_ALL_NODE_IPS="52.214.118.59 35.177.173.249 52.47.211.186 35.174.231.96 18.191.62.179 52.60.152.22 18.195.172.240"

export LOG_FILE="stability-${DATE}.log"

echo "TX_BURST_COUNT=${TX_BURST_COUNT}"
echo "INTERVAL_BETWEEN_BURSTS_MILLIS=${INTERVAL_BETWEEN_BURSTS_MILLIS}"
echo "NODE_IPS=${NODE_IPS}"
echo "VCHAIN=${VCHAIN}"
echo "GATEWAY_IP=${GATEWAY_IP}"
echo "NET_CONFIG_URL=${NET_CONFIG_URL}"

# These are unused
#echo "STRESS_TEST_NUMBER_OF_TRANSACTIONS=${STRESS_TEST_NUMBER_OF_TRANSACTIONS}"
#echo "STRESS_TEST_FAILURE_RATE=${STRESS_TEST_FAILURE_RATE}"
#echo "STRESS_TEST_TARGET_TPS=${STRESS_TEST_TARGET_TPS}"
#echo "STRESS_TEST_TRANSACTIONS_PER_MINUTE=${STRESS_TEST_TRANSACTIONS_PER_MINUTE}"
#echo "STRESS_TEST_METRICS_EVERY_NTH_TRANSACTION=${STRESS_TEST_METRICS_EVERY_NTH_TRANSACTION}"
#echo "STRESS_TEST_ALL_NODE_IPS=${STRESS_TEST_ALL_NODE_IPS}"

mkdir -p $RESULTS

echo
echo "===== STARTING TO RUN BURST=${TX_BURST_COUNT} INTERVAL=${INTERVAL_BETWEEN_BURSTS_MILLIS} NTH=${METRICS_EVERY_NTH_TX} NET_CONFIG_URL=${NET_CONFIG_URL} ====="
echo

go test ./../benchmark/... -run TestStability -tags unsafetests -timeout 100000m -count 1 -v > ${LOG_FILE} &  CMDPID=$!
echo
echo "Started process ID $CMDPID. To stop it, run:"
echo "kill $CMDPID"
echo
echo "To follow progress, run:"
echo
echo "tail -100f ${LOG_FILE}"
echo
echo
