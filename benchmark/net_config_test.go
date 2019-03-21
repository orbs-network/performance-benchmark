package benchmark

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReadUrlConfig(t *testing.T) {

	urlConfig := "https://s3.us-east-2.amazonaws.com/boyar-integrative-e2e/boyar/config.json"
	cfg, err := ReadUrlConfig(urlConfig)

	for i := 0; i < len(cfg.ValidatorNodes); i++ {
		validator := cfg.ValidatorNodes[i]
		t.Logf("%s: %s", validator.Address, validator.IP)
	}

	for i := 0; i < len(cfg.Chains); i++ {
		chain := cfg.Chains[i]
		t.Logf("%d", chain.Id)

	}

	require.NoError(t, err, "failed to read configuration")

}
