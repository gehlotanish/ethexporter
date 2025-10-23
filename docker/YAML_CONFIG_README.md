# YAML Configuration Support

ETHexporter now supports configuration via YAML files, making it much easier to manage multiple networks and addresses without setting numerous environment variables.

## Features

- **YAML Configuration**: Define networks and addresses in a single YAML file
- **Environment Variable Substitution**: Use `${VAR_NAME}` or `${VAR_NAME:-default}` syntax in YAML
- **Backward Compatibility**: Still supports environment variables if no YAML file is found
- **Chain ID Validation**: Automatic chain ID detection and validation
- **Clean Metrics**: Standardized metric names with network info in labels
- **Mixed Configuration**: Combine YAML with environment variable overrides

## Configuration File

### Default Location
- **File**: `config.yaml` (in the same directory as the binary)
- **Environment Variable**: `CONFIG_FILE` (to specify a different path)

### YAML Structure

```yaml
global:
  port: "9100"              # HTTP server port
  sleep_seconds: 15         # Update interval in seconds

networks:
  network_name:
    rpc: "https://rpc-endpoint"
    chain_id: "1"
    addresses:
      wallet_name: "0x..."
      another_wallet: "0x..."
```

## Environment Variable Substitution

YAML configuration supports environment variable substitution using `${VAR_NAME}` syntax:

### Basic Substitution
```yaml
global:
  port: "${PORT}"
  sleep_seconds: ${SLEEP_SECONDS}
```

### Default Values
```yaml
global:
  port: "${PORT:-9100}"           # Use PORT env var, default to 9100
  sleep_seconds: ${SLEEP_SECONDS:-15}  # Use SLEEP_SECONDS env var, default to 15
```

### Network Configuration with Environment Variables
```yaml
networks:
  mainnet:
    rpc: "${MAINNET_RPC:-https://eth-mainnet.public.blastapi.io}"
    prefix: "${MAINNET_PREFIX:-mainnet_}"
    chain_id: "${MAINNET_CHAIN_ID:-1}"
    addresses:
      wallet1: "${MAINNET_WALLET1:-0xb2F801913949c3eecDfc814CCc743618efF1f8c8}"
```

## Example Configuration

### With Environment Variables
```yaml
global:
  port: "${PORT:-9100}"
  sleep_seconds: ${SLEEP_SECONDS:-15}
  prefix: "${PREFIX:-}"

networks:
  mainnet:
    rpc: "${MAINNET_RPC:-https://eth-mainnet.public.blastapi.io}"
    prefix: "${MAINNET_PREFIX:-mainnet_}"
    chain_id: "${MAINNET_CHAIN_ID:-1}"
    addresses:
      wallet1: "${MAINNET_WALLET1:-0xb2F801913949c3eecDfc814CCc743618efF1f8c8}"
      wallet2: "${MAINNET_WALLET2:-0xa23D506848C30ea091B51258E00b1dC61BcD5cDb}"

  surge_hoodi:
    rpc: "${SURGE_HOODI_RPC:-https://l2-rpc.staging.surge.wtf}"
    prefix: "${SURGE_HOODI_PREFIX:-surge_hoodi_}"
    chain_id: "${SURGE_HOODI_CHAIN_ID:-763374}"
    addresses:
      surge_hoodi_wallet: "${SURGE_HOODI_WALLET:-0x3bc256069FF9af461F3e04494A3ece3f62F183fC}"
```

### Static Configuration (No Environment Variables)
```yaml
global:
  port: "9100"
  sleep_seconds: 15
  prefix: ""

networks:
  mainnet:
    rpc: "https://eth-mainnet.public.blastapi.io"
    prefix: "mainnet_"
    chain_id: "1"
    addresses:
      wallet1: "0xb2F801913949c3eecDfc814CCc743618efF1f8c8"
      wallet2: "0xa23D506848C30ea091B51258E00b1dC61BcD5cDb"
```

## Usage

### 1. Local Development
```bash
# Create config.yaml in the same directory
cp config.yaml.example config.yaml

# Edit the configuration
nano config.yaml

# Run the application
go run main.go
```

### 2. Docker Compose
```yaml
version: "3.9"
services:
  ethexporter:
    build:
      context: ..
      dockerfile: docker/Dockerfile
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    environment:
      CONFIG_FILE: "/app/config.yaml"
    ports:
      - "9100:9100"
    restart: unless-stopped
```

### 3. Docker with Custom Config
```yaml
version: "3.9"
services:
  ethexporter:
    image: ghcr.io/gehlotanish/ethexporter:latest
    volumes:
      - ./my-config.yaml:/app/config.yaml:ro
    environment:
      CONFIG_FILE: "/app/config.yaml"
    ports:
      - "9100:9100"
```

## Configuration Options

### Global Settings
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | string | "9100" | HTTP server port |
| `sleep_seconds` | integer | 15 | Update interval in seconds |
| `prefix` | string | "" | Global prefix for metrics |

### Network Settings
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `rpc` | string | ✅ | RPC endpoint URL |
| `prefix` | string | ❌ | Network-specific metric prefix |
| `chain_id` | string | ❌ | Expected chain ID (for validation) |
| `addresses` | map | ✅ | Addresses to monitor |

## Metrics Output

With YAML configuration, metrics will include network and chain_id labels:

```
mainnet_eth_balance{name="wallet1",address="0xb2F801913949c3eecDfc814CCc743618efF1f8c8",network="mainnet",chain_id="1"} 1.5
surge_hoodi_eth_balance{name="surge_hoodi_wallet",address="0x3bc256069FF9af461F3e04494A3ece3f62F183fC",network="surge_hoodi",chain_id="763374"} 999886502.8
```

## Migration from Environment Variables

### Before (Environment Variables)
```bash
NETWORK_MAINNET_RPC="https://eth-mainnet.public.blastapi.io"
NETWORK_MAINNET_PREFIX="mainnet_"
NETWORK_MAINNET_CHAIN_ID="1"
NETWORK_MAINNET_ADDR_wallet1="0xb2F801913949c3eecDfc814CCc743618efF1f8c8"
PORT="9100"
SLEEP_SECONDS="15"
```

### After (YAML Configuration)
```yaml
global:
  port: "9100"
  sleep_seconds: 15

networks:
  mainnet:
    rpc: "https://eth-mainnet.public.blastapi.io"
    prefix: "mainnet_"
    chain_id: "1"
    addresses:
      wallet1: "0xb2F801913949c3eecDfc814CCc743618efF1f8c8"
```

## Configuration Modes

ETHexporter supports multiple configuration modes:

### 1. **YAML with Environment Variables** (Recommended)
```yaml
# config.yaml
global:
  port: "${PORT:-9100}"
networks:
  mainnet:
    rpc: "${MAINNET_RPC:-https://eth-mainnet.public.blastapi.io}"
```
```bash
# Set environment variables
export MAINNET_RPC="https://custom-rpc.com"
export PORT="9101"
go run main.go
```

### 2. **Pure YAML Configuration**
```yaml
# config.yaml
global:
  port: "9100"
networks:
  mainnet:
    rpc: "https://eth-mainnet.public.blastapi.io"
```
```bash
go run main.go
```

### 3. **Environment Variables Only** (Legacy)
```bash
export NETWORK_MAINNET_RPC="https://eth-mainnet.public.blastapi.io"
export NETWORK_MAINNET_PREFIX="mainnet_"
export PORT="9100"
go run main.go
```

## Backward Compatibility

ETHexporter maintains full backward compatibility:

1. **YAML First**: If `config.yaml` exists, it will be used
2. **Environment Fallback**: If no YAML file is found, environment variables are used
3. **Mixed Mode**: You can still use environment variables to override YAML settings

## Advanced Configuration

### Custom Config File Location
```bash
CONFIG_FILE="/path/to/my-config.yaml" go run main.go
```

### Multiple Config Files
```bash
# Development
CONFIG_FILE="config-dev.yaml" go run main.go

# Production
CONFIG_FILE="config-prod.yaml" go run main.go
```

### Environment Override
```bash
# Use YAML but override port
CONFIG_FILE="config.yaml" PORT="9101" go run main.go
```

## Troubleshooting

### Config File Not Found
```
No YAML config file found, using environment variables
```
**Solution**: Ensure `config.yaml` exists in the working directory or set `CONFIG_FILE` environment variable.

### Invalid YAML Syntax
```
failed to parse YAML config: yaml: line 5: found character that cannot start any token
```
**Solution**: Check YAML syntax, ensure proper indentation and quotes.

### Chain ID Mismatch
```
Warning: Chain ID mismatch for network mainnet: expected 1, got 137
```
**Solution**: Update the `chain_id` in your YAML config to match the actual network.

### Address Validation
```
Warning: Invalid address for network mainnet: wallet1
```
**Solution**: Ensure all addresses are valid Ethereum addresses (42 characters, starting with 0x).

## Best Practices

1. **Use descriptive network names**: `mainnet`, `testnet`, `staging`
2. **Validate chain IDs**: Always specify expected chain IDs for validation
3. **Organize addresses**: Use meaningful names for addresses
4. **Version control**: Keep your config files in version control
5. **Environment-specific configs**: Use different config files for different environments
6. **Clean metrics**: Network information is automatically included in metric labels
