package benchmark

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type nodeConfiguration struct {
	Chains         []*VirtualChain  `json:"chains"`
	ValidatorNodes []*ValidatorNode `json:"network"`
}

type VirtualChainId uint32

type VirtualChain struct {
	Id         VirtualChainId
	HttpPort   int
	GossipPort int
	Config     map[string]interface{}
}

type ValidatorNode struct {
	Address string `json:"address"`
	IP      string `json:"ip"`
}

func ReadUrlConfig(configUrl string) (*nodeConfiguration, error) {
	resp, err := http.Get(configUrl)
	if err != nil {
		return nil, fmt.Errorf("could not download configuration from source: %s", err)
	}

	defer resp.Body.Close()

	input, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read configuration from source: %s", err)
	}

	return parseStringConfig(string(input))
}

func parseStringConfig(input string) (*nodeConfiguration, error) {
	var value nodeConfiguration
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		return nil, err
	}
	return &value, nil
}
