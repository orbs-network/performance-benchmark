#!/bin/sh

DATE=$(date +%Y-%m-%d-%H%M%S)
export LOG_FILE="reelect-${DATE}.log"

echo "===== STARTING TO RUN PERIODIC REELECTION ====="
echo
echo "To follow progress, run: tail -f ${LOG_FILE}"
echo

go test ./../benchmark/... -run TestPeriodicReelection -tags unsafetests -timeout 100000m -count 1 -v > ${LOG_FILE} &  CMDPID=$!
echo
echo "Started process ID $CMDPID. To stop it, run:"
echo "kill $CMDPID"
echo
echo "Tailing the log file, you can safely stop it with ^C."
echo "Process $CMDPID will continue in the background"
echo
tail -f ${LOG_FILE}

