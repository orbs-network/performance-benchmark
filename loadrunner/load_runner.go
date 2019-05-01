package main

import (
	"fmt"
	"github.com/orbs-network/performance-benchmark/benchmark"
	"os"
	"strconv"
)

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) < 4 {
		fmt.Printf("Not enough parameters: Usage: <VCHAIN> <TPS> <CFG_PATH> <IS_ASYNC>")
		os.Exit(1)
	}
	vchain, err := strconv.ParseInt(argsWithoutProg[0], 10, 0)
	if err != nil {
		panic(err)
	}
	tps, err := strconv.ParseInt(argsWithoutProg[1], 10, 0)
	if err != nil {
		panic(err)
	}
	cfgPath := argsWithoutProg[2]

	isAsync, _ := strconv.ParseBool(argsWithoutProg[3])

	loadRunner(vchain, uint64(tps), cfgPath, isAsync)

}

func loadRunner(vchain int64, tps uint64, cfgPath string, isAsync bool) {
	config := benchmark.GetLoadRunnerConfig(vchain, tps, cfgPath, isAsync)
	h := benchmark.NewHarness(config)

	benchmark.RunLoad(h, config)

}
