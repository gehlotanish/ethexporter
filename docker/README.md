# Docker usage

Build image:
```bash
docker build -t ethexporter:latest -f docker/Dockerfile .
```

Run container:
```bash
docker run --rm -p 9100:9100 \
  -e RPC=https://your-rpc-endpoint \
  -e PORT=9100 \
  -e SLEEP_SECONDS=15 \
  -e ethaddr_wallet1=0xYourAddress1 \
  -e ETHADDR_wallet2=0xYourAddress2 \
  ethexporter:latest
```

Metrics endpoint: http://localhost:9100/metrics

Notes:
- RPC and PORT are required.
- Define as many addresses as you like using `ethaddr_*` or `ETHADDR_*` envs.
