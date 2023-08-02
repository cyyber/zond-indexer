# zond-indexer

This is the indexer for the Zond POS blockchain. The features included in this indexer are:

- Indexer for indexing epoch, block, transaction, token, validator, staking and price data.
- Explorer dashboard for displaying the indexer data
- Integration of mongodb and postgresql
- Apis to fetch indexed data
- Swagger integration for api specification

### Requirements

* [Go 1.18+](https://golang.org/dl/)
* Postgresql
* Mongodb

### Getting Started

- Clone the repository - `git clone https://github.com/Prajjawalk/zond-indexer.git`
- Copy the `config.example.yaml` file and make changes according to your environment.
- Perform database schema migrations using `go run cmd/migrations/postgres.go`

### Generating swagger api spec doc
```
$ go install github.com/swaggo/swag/cmd/swag@v1.8.3 && swag init --exclude bin,_gitignore,.vscode,.idea --parseDepth 1 -g ./handlers/api.go
```

### Starting cache updater for frontend
```
$ go run cmd/frontend-data-updater/main.go --config {path_to_config}
```

### Starting go routines to export periodic validator deposits stats to db
```
$ go run cmd/ethstore-exporter/main.go --config {path_to_config}
```

### Starting indexer
```
$ go run cmd/explorer/main.go --config {path_to_config}
```

### Starting execution node metadata exporter
```
$ go run cmd/eth1indexer/main.go --config {path_to_config}
```
