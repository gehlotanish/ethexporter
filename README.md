# ethexporter

Advanced Ethereum wallet exporter that supports multiple networks, YAML configuration, and exposes per-address metrics over HTTP for Prometheus scraping.

### Build

```bash
go build
```

### Configure

#### YAML Configuration (Recommended)

Create a `config.yaml` file:

```yaml
global:
  port: "9100"
  sleep_seconds: 15

networks:
  mainnet:
    rpc: "https://eth-mainnet.public.blastapi.io"
    chain_id: "1"
    addresses:
      wallet1: "0xb2F801913949c3eecDfc814CCc743618efF1f8c8"
      wallet2: "0xa23D506848C30ea091B51258E00b1dC61BcD5cDb"

  surge_hoodi:
    rpc: "https://l2-rpc.staging.surge.wtf"
    chain_id: "763374"
    addresses:
      wallet: "0x3bc256069FF9af461F3e04494A3ece3f62F183fC"
```

#### Environment Variables (Legacy)

Set environment variables before running:

- RPC: HTTP RPC endpoint of your Ethereum L1 node
- PORT: HTTP port to serve metrics on (default unset → you must set)
- SLEEP_SECONDS: optional refresh interval in seconds (default: 15)
- ethaddr_* / ETHADDR_*: one env var per address to watch. The suffix becomes the name label.

Examples:

```bash
export RPC=http://127.0.0.1:8545
export PORT=9100
export SLEEP_SECONDS=30   # optional, default 15

# addresses (case-insensitive key prefix)
export ethaddr_treasury=0xYourAddress1
export ETHADDR_ops=0xYourAddress2

./ethexporter
```

Refresh interval: configurable via SLEEP_SECONDS (default 15 seconds).

## Features

- **Multi-Network Support**: Monitor addresses across different Ethereum networks
- **YAML Configuration**: Clean configuration file with environment variable substitution
- **Mixed Configuration**: Combine YAML config with environment variables
- **Chain ID Validation**: Automatic chain ID detection and validation
- **Clean Metrics**: Standardized metric names with network info in labels
- **Backward Compatibility**: Legacy environment variable configuration still works

### Metrics

Per-address (with network and chain_id labels):

- eth_balance{name, address, network, chain_id}
- eth_balance_pending{name, address, network, chain_id}
- eth_nonce{name, address, network, chain_id}
- eth_nonce_pending{name, address, network, chain_id}
- eth_is_contract{name, address, network, chain_id}
- eth_code_size_bytes{name, address, network, chain_id}
- eth_last_updated_unixtime{name, address, network, chain_id}

Exporter totals:

- eth_balance_total
- eth_contract_addresses_total
- eth_eoa_addresses_total
- eth_load_seconds
- eth_loaded_addresses
- eth_total_addresses

If PREFIX is set, it is prepended to every metric name (e.g., "myapp_eth_balance").

### Notes

- The exporter reads addresses only from environment variables with prefix `ethaddr_` (case-insensitive). File-based loading is not used.
- Use a reliable mainnet RPC (self-hosted Nethermind/GetH, or a provider with API key). Some public endpoints may throttle or return zeros.

### Troubleshooting

- Seeing zeros for balances/nonces/code? Verify GETH points to a working mainnet RPC and that your node is fully synced.
- Confirm live output:

```bash
curl -s http://127.0.0.1:9100/metrics | head -n 40
```

### Docker

Build:
```bash
docker build -t ethexporter:latest .
```

Run:
```bash
docker run --rm -p 9100:9100 \
  -e RPC=https://your-rpc-endpoint \
  -e PORT=9100 \
  -e SLEEP_SECONDS=15 \
  -e ethaddr_wallet1=0xYourAddress1 \
  -e ETHADDR_wallet2=0xYourAddress2 \
  ethexporter:latest
```
