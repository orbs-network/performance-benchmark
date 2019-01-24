#!/bin/bash -xe

export API_ENDPOINT=${API_ENDPOINT-http://localhost:8080}
export COMMIT=${COMMIT-master}
export RESULTS=results/${COMMIT}/$(date +%Y-%m-%dT%H:%M:%SZ)

mkdir -p $RESULTS

for profile in block goroutine heap mutex threadcreate; do
    curl -sK -v ${API_ENDPOINT}/debug/pprof/${profile} > ${RESULTS}/${profile}.out
done

curl -sK -v ${API_ENDPOINT}/metrics > ${RESULTS}/metrics.json

echo $RESULTS
