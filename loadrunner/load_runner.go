package main

import (
	"github.com/orbs-network/performance-benchmark/benchmark"
	"os"
	"strconv"
)

func main() {
	argsWithoutProg := os.Args[1:]
	vchain, err := strconv.ParseInt(argsWithoutProg[0], 10, 0)
	if err != nil {
		panic(err)
	}
	tps, err := strconv.ParseInt(argsWithoutProg[1], 10, 0)
	if err != nil {
		panic(err)
	}
	cfgPath := argsWithoutProg[2]
	if len(cfgPath) == 0 {
		cfgPath = "../config/prod-topology.json"
	}

	loadRunner(vchain, uint64(tps), cfgPath)

}

func loadRunner(vchain int64, tps uint64, cfgPath string) {
	config := benchmark.GetLoadRunnerConfig(vchain, tps, cfgPath)
	h := benchmark.NewHarness(config)

	benchmark.RunLoad(h, config)

}
