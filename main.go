package main

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	allWatching []*Watching
	port        string
	prefix      string
	eth         *ethclient.Client
	watchingMu  sync.RWMutex
	loadSeconds float64
	totalLoaded int64
)

type Watching struct {
	Name           string
	Address        string
	Balance        string
	BalancePending string
	Nonce          uint64
	NoncePending   uint64
	IsContract     bool
	CodeSize       int
	LastUpdated    int64
}

func ConnectionToGeth(url string) error {
	var err error
	eth, err = ethclient.Dial(url)
	return err
}

func GetEthBalance(address string) *big.Float {
	balance, err := eth.BalanceAt(context.TODO(), common.HexToAddress(address), nil)
	if err != nil {
		fmt.Printf("Error fetching ETH Balance for address: %v\n", address)
	}
	return ToEther(balance)
}

func UpdateAddressMetrics(w *Watching) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := common.HexToAddress(w.Address)

	var balanceStr, pendingStr string
	var nonce, pnonce uint64
	var isContract bool
	var codeSize int

	if bal, err := eth.BalanceAt(ctx, addr, nil); err == nil {
		balanceStr = ToEther(bal).String()
	}

	if pbal, err := eth.PendingBalanceAt(ctx, addr); err == nil {
		pendingStr = ToEther(pbal).String()
	}

	if n, err := eth.NonceAt(ctx, addr, nil); err == nil {
		nonce = n
	}
	if pn, err := eth.PendingNonceAt(ctx, addr); err == nil {
		pnonce = pn
	}

	if code, err := eth.CodeAt(ctx, addr, nil); err == nil {
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

	watchingMu.RLock()
	for _, v := range allWatching {
		balStr := v.Balance
		if balStr == "" {
			balStr = "0"
		}
		bal := big.NewFloat(0)
		bal.SetString(balStr)
		total.Add(total, bal)

		allOut = append(allOut, fmt.Sprintf("%veth_balance{name=\"%v\",address=\"%v\"} %v", prefix, v.Name, v.Address, balStr))
		pbalStr := v.BalancePending
		if pbalStr == "" {
			pbalStr = "0"
		}
		allOut = append(allOut, fmt.Sprintf("%veth_balance_pending{name=\"%v\",address=\"%v\"} %v", prefix, v.Name, v.Address, pbalStr))
		allOut = append(allOut, fmt.Sprintf("%veth_nonce{name=\"%v\",address=\"%v\"} %d", prefix, v.Name, v.Address, v.Nonce))
		allOut = append(allOut, fmt.Sprintf("%veth_nonce_pending{name=\"%v\",address=\"%v\"} %d", prefix, v.Name, v.Address, v.NoncePending))
		if v.IsContract {
			contracts++
			allOut = append(allOut, fmt.Sprintf("%veth_is_contract{name=\"%v\",address=\"%v\"} 1", prefix, v.Name, v.Address))
		} else {
			eoas++
			allOut = append(allOut, fmt.Sprintf("%veth_is_contract{name=\"%v\",address=\"%v\"} 0", prefix, v.Name, v.Address))
		}
		allOut = append(allOut, fmt.Sprintf("%veth_code_size_bytes{name=\"%v\",address=\"%v\"} %d", prefix, v.Name, v.Address, v.CodeSize))
		allOut = append(allOut, fmt.Sprintf("%veth_last_updated_unixtime{name=\"%v\",address=\"%v\"} %d", prefix, v.Name, v.Address, v.LastUpdated))
	}
	watchingMu.RUnlock()
	allOut = append(allOut, fmt.Sprintf("%veth_contract_addresses_total %d", prefix, contracts))
	allOut = append(allOut, fmt.Sprintf("%veth_eoa_addresses_total %d", prefix, eoas))
	allOut = append(allOut, fmt.Sprintf("%veth_load_seconds %0.2f", prefix, loadSeconds))
	allOut = append(allOut, fmt.Sprintf("%veth_loaded_addresses %v", prefix, totalLoaded))
	allOut = append(allOut, fmt.Sprintf("%veth_total_addresses %v", prefix, len(allWatching)))
	fmt.Fprintln(w, strings.Join(allOut, "\n"))
}

func OpenAddressesFromEnv(envPrefix string) error {
	if envPrefix == "" {
		envPrefix = "ethaddr_"
	}
	lowerPrefix := strings.ToLower(envPrefix)
	loaded := 0
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
			w := &Watching{
				Name:           name,
				Address:        val,
				Balance:        "0",
				BalancePending: "0",
			}
			allWatching = append(allWatching, w)
			loaded++
		}
	}
	if loaded == 0 {
		return fmt.Errorf("no addresses found in environment with prefix %q", envPrefix)
	}
	return nil
}

func main() {
	gethUrl := os.Getenv("RPC")
	port = os.Getenv("PORT")
	prefix = os.Getenv("PREFIX")

	if gethUrl == "" {
		fmt.Println("Missing required env RPC")
		os.Exit(1)
	}
	if port == "" {
		fmt.Println("Missing required env PORT")
		os.Exit(1)
	}

	sleepSeconds := 15
	if v := os.Getenv("SLEEP_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			sleepSeconds = n
		}
	}

	err := OpenAddressesFromEnv("ethaddr_")
	if err != nil {
		panic(err)
	}

	err = ConnectionToGeth(gethUrl)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			t1 := time.Now()
			fmt.Printf("Checking %v wallets...\n", len(allWatching))

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
			fmt.Printf("Finished checking %v wallets in %0.0f seconds, sleeping for %v seconds.\n", len(allWatching), loadSeconds, sleepSeconds)
			time.Sleep(time.Duration(sleepSeconds) * time.Second)
		}
	}()

	fmt.Printf("ETHexporter has started on port %v using RPC endpoint: %v\n", port, gethUrl)
	http.HandleFunc("/metrics", MetricsHttp)
	panic(http.ListenAndServe("0.0.0.0:"+port, nil))
}
