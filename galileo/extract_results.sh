#!/bin/bash -xe

# BASE_URL has the format: http://54.194.120.89/vchains/2000

export COMMIT=${COMMIT-master}
export RESULTS=${RESULTS-results/${COMMIT}/$(date +%Y-%m-%d-%H%M%S)}

mkdir -p $RESULTS

for profile in block goroutine heap mutex threadcreate; do
    echo ">>> Reading ${BASE_URL}/debug/pprof/${profile} <<<"
    curl -sK --connect-timeout=2s -v ${BASE_URL}/debug/pprof/${profile} > ${RESULTS}/${profile}.out
    echo
done

echo ">>> Reading ${BASE_URL}/metrics <<<"
curl -sK --connect-timeout=2s -v ${BASE_URL}/metrics > ${RESULTS}/metrics.json

go tool pprof -top $RESULTS/goroutine.out

echo $RESULTS