package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/yaml.v3"
)

var (
	allWatching []*Watching
	port        string
	networks    map[string]*NetworkConfig
	watchingMu  sync.RWMutex
	loadSeconds float64
	totalLoaded int64
)

type NetworkConfig struct {
	Name      string
	RPC       string
	Client    *ethclient.Client
	ChainID   *big.Int
	Addresses []*Watching
}

type YAMLConfig struct {
	Networks map[string]YAMLNetwork `yaml:"networks"`
	Global   YAMLGlobal             `yaml:"global"`
}

type YAMLNetwork struct {
	RPC       string            `yaml:"rpc"`
	ChainID   string            `yaml:"chain_id"`
	Addresses map[string]string `yaml:"addresses"`
}

type YAMLGlobal struct {
	Port         string `yaml:"port"`
	SleepSeconds int    `yaml:"sleep_seconds"`
}

type Watching struct {
	Name           string
	Address        string
	Network        string
	Balance        string
	BalancePending string
	Nonce          uint64
	NoncePending   uint64
	IsContract     bool
	CodeSize       int
	LastUpdated    int64
}

func ConnectionToGeth(url string) (*ethclient.Client, error) {
	client, err := ethclient.Dial(url)
	return client, err
}

func GetChainID(client *ethclient.Client) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}
	return chainID, nil
}

func ValidateChainID(client *ethclient.Client, expectedChainID *big.Int) error {
	actualChainID, err := GetChainID(client)
	if err != nil {
		return err
	}

	if actualChainID.Cmp(expectedChainID) != 0 {
		return fmt.Errorf("chain ID mismatch: expected %v, got %v", expectedChainID, actualChainID)
	}

	return nil
}

func expandEnvVars(content string) string {
	result := content

	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}

		end += start
		pattern := result[start+2 : end]

		var replacement string
		if strings.Contains(pattern, ":-") {
			parts := strings.SplitN(pattern, ":-", 2)
			varName := parts[0]
			defaultValue := parts[1]

			if envValue := os.Getenv(varName); envValue != "" {
				replacement = envValue
			} else {
				replacement = defaultValue
			}
		} else {
			if envValue := os.Getenv(pattern); envValue != "" {
				replacement = envValue
			} else {
				replacement = ""
			}
		}

		result = result[:start] + replacement + result[end+1:]
	}

	return result
}

func LoadYAMLConfig(configPath string) (*YAMLConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", configPath, err)
	}

	configContent := expandEnvVars(string(data))

	var config YAMLConfig
	err = yaml.Unmarshal([]byte(configContent), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %v", err)
	}

	return &config, nil
}

func InitializeNetworksFromYAML(config *YAMLConfig) error {
	networks = make(map[string]*NetworkConfig)

	for networkName, networkConfig := range config.Networks {
		var expectedChainID *big.Int
		if networkConfig.ChainID != "" {
			chainID, ok := new(big.Int).SetString(networkConfig.ChainID, 10)
			if !ok {
				return fmt.Errorf("invalid chain ID for network %s: %s", networkName, networkConfig.ChainID)
			}
			expectedChainID = chainID
		}

		client, err := ConnectionToGeth(networkConfig.RPC)
		if err != nil {
			return fmt.Errorf("failed to connect to network %s: %v", networkName, err)
		}

		actualChainID, err := GetChainID(client)
		if err != nil {
			actualChainID = nil
		} else {
			if expectedChainID != nil {
				if actualChainID.Cmp(expectedChainID) != 0 {
					fmt.Printf("Warning: Chain ID mismatch for network %s: expected %v, got %v\n", networkName, expectedChainID, actualChainID)
				}
			}
		}

		networks[networkName] = &NetworkConfig{
			Name:    networkName,
			RPC:     networkConfig.RPC,
			Client:  client,
			ChainID: actualChainID,
		}

		for addrName, addrValue := range networkConfig.Addresses {
			if common.IsHexAddress(addrValue) {
				w := &Watching{
					Name:           addrName,
					Address:        addrValue,
					Network:        networkName,
					Balance:        "0",
					BalancePending: "0",
				}
				allWatching = append(allWatching, w)
			}
		}
	}

	return nil
}

func InitializeNetworks() error {
	if networks == nil {
		networks = make(map[string]*NetworkConfig)
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := strings.TrimSpace(parts[1])

		if strings.HasPrefix(key, "NETWORK_") && strings.HasSuffix(key, "_RPC") {
			networkName := key[8 : len(key)-4]
			networkName = strings.ToLower(networkName)

			if _, exists := networks[networkName]; exists {
				continue
			}

			chainIDKey := fmt.Sprintf("NETWORK_%s_CHAIN_ID", strings.ToUpper(networkName))
			chainIDStr := os.Getenv(chainIDKey)
			var expectedChainID *big.Int
			if chainIDStr != "" {
				chainID, ok := new(big.Int).SetString(chainIDStr, 10)
				if !ok {
					return fmt.Errorf("invalid chain ID for network %s: %s", networkName, chainIDStr)
				}
				expectedChainID = chainID
			}

			client, err := ConnectionToGeth(val)
			if err != nil {
				return fmt.Errorf("failed to connect to network %s: %v", networkName, err)
			}

			actualChainID, err := GetChainID(client)
			if err != nil {
				actualChainID = nil
			} else {
				if expectedChainID != nil {
					if actualChainID.Cmp(expectedChainID) != 0 {
						fmt.Printf("Warning: Chain ID mismatch for network %s: expected %v, got %v\n", networkName, expectedChainID, actualChainID)
					}
				}
			}

			networks[networkName] = &NetworkConfig{
				Name:    networkName,
				RPC:     val,
				Client:  client,
				ChainID: actualChainID,
			}
		}
	}

	if len(networks) == 0 {
		rpc := os.Getenv("RPC")
		if rpc == "" {
			return fmt.Errorf("no networks configured. Set NETWORK_<name>_RPC environment variables or RPC for backward compatibility")
		}

		client, err := ConnectionToGeth(rpc)
		if err != nil {
			return err
		}

		networks["default"] = &NetworkConfig{
			Name:   "default",
			RPC:    rpc,
			Client: client,
		}
	}

	return nil
}

func GetEthBalance(client *ethclient.Client, address string) *big.Float {
	balance, err := client.BalanceAt(context.TODO(), common.HexToAddress(address), nil)
	if err != nil {
		return big.NewFloat(0)
	}
	return ToEther(balance)
}

func UpdateAddressMetrics(w *Watching) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := common.HexToAddress(w.Address)

	network, exists := networks[w.Network]
	if !exists {
		return
	}

	client := network.Client

	var balanceStr, pendingStr string
	var nonce, pnonce uint64
	var isContract bool
	var codeSize int

	if bal, err := client.BalanceAt(ctx, addr, nil); err == nil {
		balanceStr = ToEther(bal).String()
	}

	if pbal, err := client.PendingBalanceAt(ctx, addr); err == nil {
		pendingStr = ToEther(pbal).String()
	}

	if n, err := client.NonceAt(ctx, addr, nil); err == nil {
		nonce = n
	}
	if pn, err := client.PendingNonceAt(ctx, addr); err == nil {
		pnonce = pn
	}

	if code, err := client.CodeAt(ctx, addr, nil); err == nil {
		codeSize = len(code)
		isContract = len(code) > 0
	}

	watchingMu.Lock()
	w.Balance = balanceStr
	w.BalancePending = pendingStr
	w.Nonce = nonce
	w.NoncePending = pnonce
	w.IsContract = isContract
	w.CodeSize = codeSize
	w.LastUpdated = time.Now().Unix()
	watchingMu.Unlock()
}

func ToEther(o *big.Int) *big.Float {
	pul, int := big.NewFloat(0), big.NewFloat(0)
	int.SetInt(o)
	pul.Mul(big.NewFloat(0.000000000000000001), int)
	return pul
}

func MetricsHttp(w http.ResponseWriter, r *http.Request) {
	var allOut []string
	total := big.NewFloat(0)
	contracts := 0
	eoas := 0
	networkStats := make(map[string]int)

	watchingMu.RLock()
	for _, v := range allWatching {
		balStr := v.Balance
		if balStr == "" {
			balStr = "0"
		}
		bal := big.NewFloat(0)
		bal.SetString(balStr)
		total.Add(total, bal)

		chainID := "unknown"
		if network, exists := networks[v.Network]; exists && network.ChainID != nil {
			chainID = network.ChainID.String()
		}

		allOut = append(allOut, fmt.Sprintf("eth_balance{name=\"%v\",address=\"%v\",network=\"%v\",chain_id=\"%v\"} %v", v.Name, v.Address, v.Network, chainID, balStr))
		pbalStr := v.BalancePending
		if pbalStr == "" {
			pbalStr = "0"
		}
		allOut = append(allOut, fmt.Sprintf("eth_balance_pending{name=\"%v\",address=\"%v\",network=\"%v\",chain_id=\"%v\"} %v", v.Name, v.Address, v.Network, chainID, pbalStr))
		allOut = append(allOut, fmt.Sprintf("eth_nonce{name=\"%v\",address=\"%v\",network=\"%v\",chain_id=\"%v\"} %d", v.Name, v.Address, v.Network, chainID, v.Nonce))
		allOut = append(allOut, fmt.Sprintf("eth_nonce_pending{name=\"%v\",address=\"%v\",network=\"%v\",chain_id=\"%v\"} %d", v.Name, v.Address, v.Network, chainID, v.NoncePending))
		if v.IsContract {
			contracts++
			allOut = append(allOut, fmt.Sprintf("eth_is_contract{name=\"%v\",address=\"%v\",network=\"%v\",chain_id=\"%v\"} 1", v.Name, v.Address, v.Network, chainID))
		} else {
			eoas++
			allOut = append(allOut, fmt.Sprintf("eth_is_contract{name=\"%v\",address=\"%v\",network=\"%v\",chain_id=\"%v\"} 0", v.Name, v.Address, v.Network, chainID))
		}
		allOut = append(allOut, fmt.Sprintf("eth_code_size_bytes{name=\"%v\",address=\"%v\",network=\"%v\",chain_id=\"%v\"} %d", v.Name, v.Address, v.Network, chainID, v.CodeSize))
		allOut = append(allOut, fmt.Sprintf("eth_last_updated_unixtime{name=\"%v\",address=\"%v\",network=\"%v\",chain_id=\"%v\"} %d", v.Name, v.Address, v.Network, chainID, v.LastUpdated))

		networkStats[v.Network]++
	}
	watchingMu.RUnlock()
	for networkName, count := range networkStats {
		chainID := "unknown"
		if network, exists := networks[networkName]; exists && network.ChainID != nil {
			chainID = network.ChainID.String()
		}
		allOut = append(allOut, fmt.Sprintf("eth_addresses_total{network=\"%v\",chain_id=\"%v\"} %d", networkName, chainID, count))
	}

	allOut = append(allOut, fmt.Sprintf("eth_contract_addresses_total %d", contracts))
	allOut = append(allOut, fmt.Sprintf("eth_eoa_addresses_total %d", eoas))
	allOut = append(allOut, fmt.Sprintf("eth_load_seconds %0.2f", loadSeconds))
	allOut = append(allOut, fmt.Sprintf("eth_loaded_addresses %v", totalLoaded))
	allOut = append(allOut, fmt.Sprintf("eth_total_addresses %v", len(allWatching)))
	fmt.Fprintln(w, strings.Join(allOut, "\n"))
}

func OpenAddressesFromEnv() error {
	loaded := 0

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := strings.TrimSpace(parts[1])

		if strings.HasPrefix(key, "NETWORK_") && strings.Contains(key, "_ADDR_") {
			addrIndex := strings.Index(key, "_ADDR_")
			if addrIndex > 0 {
				networkName := strings.ToLower(key[8:addrIndex])
				addrName := key[addrIndex+6:]

				if common.IsHexAddress(val) {
					w := &Watching{
						Name:           addrName,
						Address:        val,
						Network:        networkName,
						Balance:        "0",
						BalancePending: "0",
					}
					allWatching = append(allWatching, w)
					loaded++
				}
			}
		}
	}

	if loaded == 0 {
		envPrefix := "ethaddr_"
		lowerPrefix := strings.ToLower(envPrefix)
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := parts[0]
			val := strings.TrimSpace(parts[1])
			if !strings.HasPrefix(strings.ToLower(key), lowerPrefix) {
				continue
			}
			name := key[len(envPrefix):]
			if common.IsHexAddress(val) {
				networkName := "default"
				if len(networks) > 0 {
					for name := range networks {
						networkName = name
						break
					}
				}

				w := &Watching{
					Name:           name,
					Address:        val,
					Network:        networkName,
					Balance:        "0",
					BalancePending: "0",
				}
				allWatching = append(allWatching, w)
				loaded++
			}
		}
	}

	if loaded == 0 {
		return fmt.Errorf("no addresses found in environment. Use NETWORK_<name>_ADDR_<name> or ethaddr_<name> format")
	}
	return nil
}

func main() {
	var sleepSeconds int

	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		configPath = "config.yaml"
	}

	if _, err := os.Stat(configPath); err == nil {
		config, err := LoadYAMLConfig(configPath)
		if err != nil {
			panic(fmt.Errorf("failed to load YAML config: %v", err))
		}

		port = config.Global.Port
		if port == "" {
			port = "9100"
		}

		sleepSeconds = config.Global.SleepSeconds
		if sleepSeconds == 0 {
			sleepSeconds = 15
		}

		err = InitializeNetworksFromYAML(config)
		if err != nil {
			panic(err)
		}

		err = InitializeNetworks()
		if err != nil {
			fmt.Printf("Warning: Failed to load additional networks from environment: %v\n", err)
		}

		err = OpenAddressesFromEnv()
		if err != nil {
			fmt.Printf("Warning: Failed to load additional addresses from environment: %v\n", err)
		}
	} else {
		port = os.Getenv("PORT")

		if port == "" {
			fmt.Println("Missing required env PORT")
			os.Exit(1)
		}

		sleepSeconds = 15
		if v := os.Getenv("SLEEP_SECONDS"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				sleepSeconds = n
			}
		}

		err := InitializeNetworks()
		if err != nil {
			panic(err)
		}

		err = OpenAddressesFromEnv()
		if err != nil {
			panic(err)
		}
	}

	go func() {
		for {
			t1 := time.Now()

			concurrency := 8
			sem := make(chan struct{}, concurrency)
			var wg sync.WaitGroup

			watchingMu.RLock()
			snapshot := make([]*Watching, len(allWatching))
			copy(snapshot, allWatching)
			watchingMu.RUnlock()

			for _, v := range snapshot {
				wg.Add(1)
				sem <- struct{}{}
				go func(wi *Watching) {
					defer wg.Done()
					defer func() { <-sem }()
					UpdateAddressMetrics(wi)
				}(v)
			}

			wg.Wait()

			loadSeconds = time.Since(t1).Seconds()
			totalLoaded = int64(len(allWatching))
			time.Sleep(time.Duration(sleepSeconds) * time.Second)
		}
	}()

	fmt.Printf("ETHexporter has started on port %v\n", port)

	http.HandleFunc("/metrics", MetricsHttp)
	panic(http.ListenAndServe("0.0.0.0:"+port, nil))
}
