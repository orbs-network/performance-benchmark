#!/bin/bash -xe

export API_ENDPOINT=${API_ENDPOINT-http://localhost:8080}
export BASE_URL=${BASE_URL-http://localhost:8080}
export COMMIT=${COMMIT-master}
export RESULTS=${RESULTS-results/${COMMIT}/$(date +%Y-%m-%d-%H%M%S)}

export STRESS_TEST_NUMBER_OF_TRANSACTIONS=100

echo "STRESS_TEST_NUMBER_OF_TRANSACTIONS=${STRESS_TEST_NUMBER_OF_TRANSACTIONS}"
echo "STRESS_TEST_FAILURE_RATE=${STRESS_TEST_FAILURE_RATE}"
echo "STRESS_TEST_TARGET_TPS=${STRESS_TEST_TARGET_TPS}"
echo "STRESS_TEST_TRANSACTIONS_PER_MINUTE=${STRESS_TEST_TRANSACTIONS_PER_MINUTE}"
echo "STRESS_TEST_METRICS_EVERY_NTH_TRANSACTION=${STRESS_TEST_METRICS_EVERY_NTH_TRANSACTION}"

echo "===== START ====="
go test ./../benchmark/... -count 1 -v
echo "===== STOP ====="

mkdir -p $RESULTS

for profile in block goroutine heap mutex threadcreate; do
    curl -sK --connect-timeout 2s -v ${BASE_URL}/debug/pprof/${profile} > ${RESULTS}/${profile}.out
done

curl -sK --connect-timeout 2s -v ${BASE_URL}/metrics > ${RESULTS}/metrics.json

go tool pprof -top $RESULTS/goroutine.out

echo $RESULTS