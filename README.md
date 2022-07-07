# TRON Blocks Parser
Tron parser for [Kattana](https://app.kattana.io), requirements:
* One block in one moment of time
* Horizontally scalable

### How it works?
Consume messages with block metainformation from kafka feed **tron_blocks**, parse block,
on success send it to **parser.sys.parsed**, on failure - send to **failed_blocks**.

### Dependencies
* Redis
* Docker
* Golang 1.18
* Kafka

## Configuration
**Development:**
1) Copy .env.example -> .env
2) Setup ENV=development

**Production:**
1) Setup ENV variables from .env.example

## Run
```
go build -o ./app ./cmd/main.go
./app
```