# ethexporter

Minimal Ethereum L1 wallet exporter that exposes per-address metrics over HTTP for Prometheus scraping.

### Build

```bash
go build
```

### Configure

Set environment variables before running:

- RPC: HTTP RPC endpoint of your Ethereum L1 node
- PORT: HTTP port to serve metrics on (default unset → you must set)
- PREFIX: optional metric name prefix (e.g., "myapp_")
- SLEEP_SECONDS: optional refresh interval in seconds (default: 15)
- ethaddr_* / ETHADDR_*: one env var per address to watch. The suffix becomes the name label.

Examples:

```bash
export RPC=http://127.0.0.1:8545
export PORT=9100
export PREFIX=""
export SLEEP_SECONDS=30   # optional, default 15

# addresses (case-insensitive key prefix)
export ethaddr_treasury=0xYourAddress1
export ETHADDR_ops=0xYourAddress2

# choose which metrics to fetch (optional)

./ethexporter
```

Refresh interval: configurable via SLEEP_SECONDS (default 15 seconds).

### Metrics

Per-address:

- eth_balance{name, address}
- eth_balance_pending{name, address}
- eth_nonce{name, address}
- eth_nonce_pending{name, address}
- eth_is_contract{name, address}
- eth_code_size_bytes{name, address}
- eth_last_updated_unixtime{name, address}

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
