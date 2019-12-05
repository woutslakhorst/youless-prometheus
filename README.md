### Install Prometheus

```
cd .. && go get github.com/prometheus/prometheus/cmd/...
```

### Run locally

```
go run main.go
prometheus --config.file=local_config.yml
```
