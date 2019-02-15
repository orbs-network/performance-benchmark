#!/bin/bash -xe

export COMMIT=${COMMIT-master}
export RESULTS=${RESULTS-results/${COMMIT}/$(date +%Y-%m-%d-%H%M%S)}

for profile in block goroutine heap mutex threadcreate; do
    curl -sK --connect-timeout 2s -v ${BASE_URL}/debug/pprof/${profile} > ${RESULTS}/${profile}.out
done

curl -sK --connect-timeout 2s -v ${BASE_URL}/metrics > ${RESULTS}/metrics.json

go tool pprof -top $RESULTS/goroutine.out

echo $RESULTS