#!/bin/bash -xe

export API_ENDPOINT=${API_ENDPOINT-http://localhost:8080}
export BASE_URL=${BASE_URL-http://localhost:8080}
export COMMIT=${COMMIT-master}
export RESULTS=${RESULTS-results/${COMMIT}/$(date +%Y-%m-%d-%H%M%S)}

go test ./../benchmark/... -count 1

mkdir -p $RESULTS

for profile in block goroutine heap mutex threadcreate; do
    curl -sK --connect-timeout 2s -v ${BASE_URL}/debug/pprof/${profile} > ${RESULTS}/${profile}.out
done

curl -sK --connect-timeout 2s -v ${BASE_URL}/metrics > ${RESULTS}/metrics.json

echo $RESULTS
