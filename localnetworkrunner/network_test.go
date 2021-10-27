package localnetworkrunner

import (
	_ "embed"
	"encoding/json"
	"os"
	"testing"
	"time"
    "io/ioutil"
    "errors"
    "fmt"

	"github.com/ava-labs/avalanche-network-runner-local/networkrunner"
	"github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
)

func TestWrongNetworkConfigs(t *testing.T) {
    tests := []struct{
        networkConfigPath string
        expectedError error
    } {
        {
            networkConfigPath: "network_configs/empty_config.json",
            expectedError: errors.New("couldn't unmarshall network config json: unexpected end of JSON input"),
        },
    }
    for _, tt := range tests {
        givenErr := networkStartWaitStop(tt.networkConfigPath)
        assert.Equal(t, givenErr, tt.expectedError)
    }
}

func TestBasicNetwork(t *testing.T) {
    networkConfigPath := "network_configs/basic_network.json"
    if err := networkStartWaitStop(networkConfigPath); err != nil {
        t.Fatal(err)
    }
}

func networkStartWaitStop(networkConfigPath string) error {
	binMap, err := getBinMap()
    if err != nil {
        return err
    }
    networkConfigJSON, err := readNetworkConfigJSON(networkConfigPath)
	if err != nil {
        return err
	}
	networkConfig, err := getNetworkConfig(networkConfigJSON)
    if err != nil {
        return err
    }
	net, err := startNetwork(binMap, networkConfig)
    if err != nil {
        return err
    }
    if err := awaitNetwork(net); err != nil {
        return err
    }
    if err := stopNetwork(net); err != nil {
        return err
    }
    return nil
}

func getBinMap() (map[int]string, error) {
	envVarName := "AVALANCHEGO_PATH"
	avalanchegoPath, ok := os.LookupEnv(envVarName)
	if !ok {
		return nil, errors.New(fmt.Sprintf("must define env var %s", envVarName))
	}
	envVarName = "BYZANTINE_PATH"
	byzantinePath, ok := os.LookupEnv(envVarName)
	if !ok {
		return nil, errors.New(fmt.Sprintf("must define env var %s", envVarName))
	}
	binMap := map[int]string{
		networkrunner.AVALANCHEGO: avalanchegoPath,
		networkrunner.BYZANTINE:   byzantinePath,
	}
	return binMap, nil
}

func readNetworkConfigJSON(networkConfigPath string) ([]byte, error) {
    networkConfigJSON, err := ioutil.ReadFile(networkConfigPath)
	if err != nil {
        return nil, errors.New(fmt.Sprintf("couldn't read network config file %s: %s", networkConfigPath, err))
	}
    return networkConfigJSON, nil
}

func getNetworkConfig(networkConfigJSON []byte) (*networkrunner.NetworkConfig, error) {
	networkConfig := networkrunner.NetworkConfig{}
	if err := json.Unmarshal(networkConfigJSON, &networkConfig); err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't unmarshall network config json: %s", err))
	}
	return &networkConfig, nil
}

func startNetwork(binMap map[int]string, networkConfig *networkrunner.NetworkConfig) (networkrunner.Network, error) {
	logger := logrus.New()
	var net networkrunner.Network
	net, err := NewNetwork(*networkConfig, binMap, logger)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("couldn't create network: %s", err))
	}
	return net, nil
}

func awaitNetwork(net networkrunner.Network) error {
	timeoutCh := make(chan struct{})
	go func() {
		time.Sleep(5 * time.Minute)
		timeoutCh <- struct{}{}
	}()
	readyCh, errorCh := net.Ready()
	select {
	case <-readyCh:
		break
	case err := <-errorCh:
		return err
	case <-timeoutCh:
		return errors.New("network startup timeout")
	}
    return nil
}

func stopNetwork(net networkrunner.Network) error {
	err := net.Stop()
	if err != nil {
		return errors.New(fmt.Sprintf("couldn't cleanly stop network: %s", err))
	}
    return nil
}
