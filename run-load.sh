#!/bin/bash

# Config for stability network
export INTERVAL_BETWEEN_BURSTS_MILLIS=${INTERVAL_BETWEEN_BURSTS_MILLIS-60000}
export METRICS_EVERY_NTH_TX=${METRICS_EVERY_NTH_TX-5000}

if [[ $# -ne 3 ]] ; then
    echo
    echo "Usage: [VCHAIN] [TPS] [CFG_PATH]"
    echo
    exit 1
fi

export COMMIT=${COMMIT-master}
DATE=$(date +%Y-%m-%d-%H%M%S)
export RESULTS=${RESULTS-results/$COMMIT/$DATE}
export LOG_FILE="load-${VCHAIN}-${DATE}.log"

mkdir -p ${RESULTS}

echo
echo "===== STARTING TO RUN VCHAIN=${1} TPS${2} PRINT_METRICS_EVERY_NTH_TX=${METRICS_EVERY_NTH_TX} CFG_PATH=${3} ====="
echo

CMD="go run loadrunner/load_runner.go $1 $2 $3"

echo "CMD=${CMD}"

${CMD} > ${LOG_FILE} &  CMDPID=$!
echo
echo "Started process ID $CMDPID. To stop it, run:"
echo "kill $CMDPID"
echo
echo "To follow progress, run:"
echo
echo "tail -100f ${LOG_FILE}"
echo
echo
