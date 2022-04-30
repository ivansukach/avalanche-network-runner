package config

import (
	"embed"
	"fmt"
	"github.com/ava-labs/avalanche-network-runner/network"
	"github.com/ava-labs/avalanche-network-runner/network/node"
	"io/fs"
)

var (
	//go:embed default
	embeddedDefaultNetworkConfigDir embed.FS
)

// populate default network config from embedded default directory
func BuildConfig(amountOfNodes int) network.Config {
	configsDir, err := fs.Sub(embeddedDefaultNetworkConfigDir, "default")
	if err != nil {
		panic(err)
	}

	networkConfig := network.Config{
		Name:        fmt.Sprintf("Network, that consist of %d timestampVM node(-s)", amountOfNodes),
		NodeConfigs: make([]node.Config, amountOfNodes),
		LogLevel:    "INFO",
	}

	genesis, err := fs.ReadFile(configsDir, "genesis.json")
	if err != nil {
		panic(err)
	}
	networkConfig.Genesis = string(genesis)

	for i := 0; i < len(networkConfig.NodeConfigs); i++ {
		configFile, err := fs.ReadFile(configsDir, fmt.Sprintf("node%d/config.json", i))
		if err != nil {
			panic(err)
		}
		networkConfig.NodeConfigs[i].ConfigFile = string(configFile)
		stakingKey, err := fs.ReadFile(configsDir, fmt.Sprintf("node%d/staking.key", i))
		if err != nil {
			panic(err)
		}
		networkConfig.NodeConfigs[i].StakingKey = string(stakingKey)
		stakingCert, err := fs.ReadFile(configsDir, fmt.Sprintf("node%d/staking.crt", i))
		if err != nil {
			panic(err)
		}
		cChainConfig, err := fs.ReadFile(configsDir, fmt.Sprintf("node%d/cchain_config.json", i))
		if err != nil {
			panic(err)
		}
		networkConfig.NodeConfigs[i].CChainConfigFile = string(cChainConfig)
		networkConfig.NodeConfigs[i].StakingCert = string(stakingCert)
		networkConfig.NodeConfigs[i].IsBeacon = true
	}
	return networkConfig
}
