# Docker Usage

## Quick Start with Docker Compose (Recommended)

The easiest way to run ethexporter with mixed configuration:

```bash
# Start the application
docker-compose up -d

# Check metrics
curl http://localhost:9100/metrics
```

## Configuration

### Mixed Configuration (Recommended)

**RPC endpoints and Chain IDs** → Environment variables in `docker-compose.yaml`
**Wallet addresses** → YAML config file

#### 1. Update `docker-compose.yaml`:

```yaml
version: "3.9"
services:
  ethexporter:
    image: ghcr.io/gehlotanish/ethexporter:latest
    platform: linux/amd64
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    environment:
      CONFIG_FILE: "/app/config.yaml"

      # Network RPC and Chain ID configuration
      NETWORK_MAINNET_RPC: "https://eth-mainnet.public.blastapi.io"
      NETWORK_MAINNET_CHAIN_ID: "1"

      NETWORK_HOODI_RPC: "https://l2-rpc.staging.surge.wtf"
      NETWORK_HOODI_CHAIN_ID: "763374"

      NETWORK_HOODI_STAGING_RPC: "https://l2-rpc.hoodi.surge.wtf"
      NETWORK_HOODI_STAGING_CHAIN_ID: "763375"
    ports:
      - "9100:9100"
    restart: unless-stopped
```

#### 2. Create `config.yaml`:

```yaml
global:
  port: "${PORT:-9100}"
  sleep_seconds: ${SLEEP_SECONDS:-15}

networks:
  mainnet:
    addresses:
      wallet1: "0xb2F801913949c3eecDfc814CCc743618efF1f8c8"
      wallet2: "0xa23D506848C30ea091B51258E00b1dC61BcD5cDb"

  hoodi:
    addresses:
      wallet: "0x3bc256069FF9af461F3e04494A3ece3f62F183fC"

  hoodi_staging:
    addresses:
      wallet_surge: "0x3bc256069FF9af461F3e04494A3ece3f62F183fC"
```

## Manual Docker Run

### Build Image:
```bash
docker build -t ethexporter:latest -f docker/Dockerfile .
```

### Run with Mixed Configuration:
```bash
docker run --rm -p 9100:9100 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -e CONFIG_FILE="/app/config.yaml" \
  -e NETWORK_MAINNET_RPC="https://eth-mainnet.public.blastapi.io" \
  -e NETWORK_MAINNET_CHAIN_ID="1" \
  -e NETWORK_HOODI_RPC="https://l2-rpc.staging.surge.wtf" \
  -e NETWORK_HOODI_CHAIN_ID="763374" \
  ethexporter:latest
```

### Legacy Run (Environment Variables Only):
```bash
docker run --rm -p 9100:9100 \
  -e RPC=https://your-rpc-endpoint \
  -e PORT=9100 \
  -e SLEEP_SECONDS=15 \
  -e ethaddr_wallet1=0xYourAddress1 \
  -e ETHADDR_wallet2=0xYourAddress2 \
  ethexporter:latest
```

## Metrics

Access metrics at: http://localhost:9100/metrics

## Notes

- **Mixed Configuration**: RPC endpoints via environment variables, addresses via YAML
- **Environment Variables**: Required for RPC endpoints and Chain IDs
- **YAML Config**: Required for wallet addresses
- **Backward Compatibility**: Legacy environment variable configuration still works
