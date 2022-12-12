SHELL := /bin/sh

.PHONY: help
help:
	@echo "build\tサーバーを（リ）ビルドします。"
	@echo "run\tサーバーを起動します。"
	@echo "stop\tサーバーを停止します。"
	@echo "down\tサーバーを停止し、データを削除します。"
	@echo "lint\tコードの静的解析を行います。"
	@echo "test\tユニットテストを実行します。"
	@echo "doc\tドキュメントを開きます。"
	@echo "help\tこのヘルプを表示します。"

.PHONY: build
build:
	docker compose build

.PHONY: run
run:
	@echo "Docker コンテナを起動します。"
	docker compose up -d

.PHONY: stop
stop:
	@echo "Docker コンテナを停止します。"
	docker compose stop

.PHONY: down
down:
	@echo "Docker コンテナを削除します。"
	docker compose down

.PHONY: lint
lint:
	go vet ./...
	staticcheck ./...

.PHONY: test
test: run
	go test ./...

.PHONY: doc
doc:
	@echo "ブラウザで http://localhost:6060 を開いてください。"
	godoc -http=:6060
