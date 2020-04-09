package config

import (
	"encoding/json"
	"io/ioutil"
)

type MinerConfig struct {
	MinerID                    string
	BlocksProcessingBufferSize int
	OpsProcessingBufferSize    int
	OpFloodingBufferSize       int
	BlockFloodingBufferSize    int
	BlockfsPort                int
	CoordinatorConfig          CoordinatorConfig
	BlockchainConfig           BlockchainConfig
	FlooderConfig              FlooderConfig
}

type BlockchainConfig struct {
	GenesisHash    string
	OpDifficulty   int
	NoopDifficulty int
}

type FlooderConfig struct {
	PeerPort                  int
	PeerOpsFloodBufferSize    int
	PeerBlocksFloodBufferSize int
	HeatbeatRetryMax          int
	Peers                     []PeerConfig
}

type CoordinatorConfig struct {
	MaxOpsQueueSize        int
	RetryBlockQueueMaxSize int
	WaitOpsTimeoutMilli    int
}

type PeerConfig struct {
	MinerID string
	Address string
}

func ParseMinerConfig(configPath string) (*MinerConfig, error) {
	rawConfig, fileerr := ioutil.ReadFile(configPath)
	if fileerr != nil {
		return nil, fileerr
	}
	var minerConfig MinerConfig
	jsonerr := json.Unmarshal(rawConfig, &minerConfig)
	if jsonerr != nil {
		return nil, jsonerr
	}
	return &minerConfig, nil
}

func ConvertPeerConfigsToMap(peers []PeerConfig) map[string]string {
	result := make(map[string]string)
	for _, peer := range peers {
		result[peer.MinerID] = peer.Address
	}
	return result
}
