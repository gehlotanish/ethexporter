# YAML Configuration Support

ETHexporter supports flexible configuration via YAML files with mixed configuration support, making it easy to manage multiple networks and addresses.

## Features

- **Mixed Configuration**: RPC endpoints via environment variables, addresses via YAML
- **YAML Configuration**: Define addresses in a single YAML file
- **Environment Variable Substitution**: Use `${VAR_NAME}` or `${VAR_NAME:-default}` syntax in YAML
- **Backward Compatibility**: Still supports environment variables if no YAML file is found
- **Chain ID Validation**: Automatic chain ID detection and validation
- **Clean Metrics**: Standardized metric names with network info in labels
- **Flexible Setup**: Choose between mixed, pure YAML, or environment-only configuration

## Configuration File

### Default Location
- **File**: `config.yaml` (in the same directory as the binary)
- **Environment Variable**: `CONFIG_FILE` (to specify a different path)

### YAML Structure

#### Mixed Configuration (Recommended)
```yaml
global:
  port: "${PORT:-9100}"              # HTTP server port
  sleep_seconds: ${SLEEP_SECONDS:-15} # Update interval in seconds

networks:
  network_name:
    addresses:
      wallet_name: "0x..."
      another_wallet: "0x..."
```

#### Pure YAML Configuration
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

### Mixed Configuration (Recommended)

**YAML Config (config.yaml):**
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

**Environment Variables (docker-compose.yaml or shell):**
```bash
# Network RPC and Chain ID configuration
NETWORK_MAINNET_RPC="https://eth-mainnet.public.blastapi.io"
NETWORK_MAINNET_CHAIN_ID="1"

NETWORK_HOODI_RPC="https://l2-rpc.staging.surge.wtf"
NETWORK_HOODI_CHAIN_ID="763374"

NETWORK_HOODI_STAGING_RPC="https://l2-rpc.hoodi.surge.wtf"
NETWORK_HOODI_STAGING_CHAIN_ID="763375"
```

### Pure YAML Configuration
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

### 2. Docker Compose (Mixed Configuration)
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

      # Network RPC and Chain ID configuration via environment variables
      # Wallet addresses are configured in config.yaml

      # Mainnet
      NETWORK_MAINNET_RPC: "https://eth-mainnet.public.blastapi.io"
      NETWORK_MAINNET_CHAIN_ID: "1"

      # Surge Hoodi
      NETWORK_HOODI_RPC: "https://l2-rpc.staging.surge.wtf"
      NETWORK_HOODI_CHAIN_ID: "763374"

      # Surge Hoodi Staging
      NETWORK_HOODI_STAGING_RPC: "https://l2-rpc.hoodi.surge.wtf"
      NETWORK_HOODI_STAGING_CHAIN_ID: "763375"
    ports:
      - "9100:9100"
    restart: unless-stopped
```

### 3. Docker with Pure YAML Configuration
```yaml
version: "3.9"
services:
  ethexporter:
    image: ghcr.io/gehlotanish/ethexporter:latest
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    environment:
      CONFIG_FILE: "/app/config.yaml"
    ports:
      - "9100:9100"
    restart: unless-stopped
```

## Configuration Options

### Global Settings
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | string | "9100" | HTTP server port |
| `sleep_seconds` | integer | 15 | Update interval in seconds |

### Network Settings (Mixed Configuration)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `addresses` | map | ✅ | Addresses to monitor |

### Network Settings (Pure YAML Configuration)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `rpc` | string | ✅ | RPC endpoint URL |
| `chain_id` | string | ❌ | Expected chain ID (for validation) |
| `addresses` | map | ✅ | Addresses to monitor |

## Metrics Output

With YAML configuration, metrics will include network and chain_id labels:

```
eth_balance{name="wallet1",address="0xb2F801913949c3eecDfc814CCc743618efF1f8c8",network="mainnet",chain_id="1"} 1.5
eth_balance{name="wallet",address="0x3bc256069FF9af461F3e04494A3ece3f62F183fC",network="hoodi",chain_id="763374"} 10
eth_balance{name="wallet_surge",address="0x3bc256069FF9af461F3e04494A3ece3f62F183fC",network="hoodi_staging",chain_id="763375"} 1003.779903
```

## Migration from Environment Variables

### Before (Environment Variables Only)
```bash
NETWORK_MAINNET_RPC="https://eth-mainnet.public.blastapi.io"
NETWORK_MAINNET_CHAIN_ID="1"
NETWORK_MAINNET_ADDR_wallet1="0xb2F801913949c3eecDfc814CCc743618efF1f8c8"
PORT="9100"
SLEEP_SECONDS="15"
```

### After (Mixed Configuration - Recommended)
**YAML Config (config.yaml):**
```yaml
global:
  port: "${PORT:-9100}"
  sleep_seconds: ${SLEEP_SECONDS:-15}

networks:
  mainnet:
    addresses:
      wallet1: "0xb2F801913949c3eecDfc814CCc743618efF1f8c8"
```

**Environment Variables:**
```bash
NETWORK_MAINNET_RPC="https://eth-mainnet.public.blastapi.io"
NETWORK_MAINNET_CHAIN_ID="1"
```

### After (Pure YAML Configuration)
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
```

## Configuration Modes

ETHexporter supports multiple configuration modes:

### 1. **Mixed Configuration** (Recommended)
**RPC endpoints and Chain IDs** → Environment variables
**Wallet addresses** → YAML config

```yaml
# config.yaml
global:
  port: "${PORT:-9100}"
  sleep_seconds: ${SLEEP_SECONDS:-15}

networks:
  mainnet:
    addresses:
      wallet1: "0xb2F801913949c3eecDfc814CCc743618efF1f8c8"
```
```bash
# Set environment variables
export NETWORK_MAINNET_RPC="https://eth-mainnet.public.blastapi.io"
export NETWORK_MAINNET_CHAIN_ID="1"
export CONFIG_FILE="config.yaml"
go run main.go
```

### 2. **Pure YAML Configuration**
```yaml
# config.yaml
global:
  port: "9100"
  sleep_seconds: 15

networks:
  mainnet:
    rpc: "https://eth-mainnet.public.blastapi.io"
    chain_id: "1"
    addresses:
      wallet1: "0xb2F801913949c3eecDfc814CCc743618efF1f8c8"
```
```bash
go run main.go
```

### 3. **Environment Variables Only** (Legacy)
```bash
export NETWORK_MAINNET_RPC="https://eth-mainnet.public.blastapi.io"
export NETWORK_MAINNET_CHAIN_ID="1"
export NETWORK_MAINNET_ADDR_wallet1="0xb2F801913949c3eecDfc814CCc743618efF1f8c8"
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

1. **Use descriptive network names**: `mainnet`, `hoodi`, `hoodi_staging`
2. **Mixed Configuration**: Use environment variables for RPC endpoints, YAML for addresses
3. **Validate chain IDs**: Always specify expected chain IDs for validation
4. **Organize addresses**: Use meaningful names for addresses
5. **Version control**: Keep your config files in version control
6. **Environment-specific configs**: Use different config files for different environments
7. **Clean metrics**: Network information is automatically included in metric labels
8. **Docker Compose**: Use docker-compose.yaml for easy deployment with mixed configuration
