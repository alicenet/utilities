REGISTRY ?= us-central1-docker.pkg.dev/mn-test-298216/alicenet
MIGRATION_SOURCE ?= file://internal/migrations
SPANNER_DATABASE ?= projects/mn-test-298216/instances/alicenet/databases/indexer

.PHONY: all
all: setup generate format lint test build

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test -v -covermode=atomic -race ./...

.PHONY: lint
lint:
	golangci-lint run
	buf lint
	# buf breaking --against '.git#branch=main' # Not implemented yet

.PHONY: format
format:
	buf format -w
	npx prettier --write .
	golangci-lint run -E gci,godot,gofumpt,misspell,whitespace --fix
	go mod tidy -v

.PHONY: generate
generate:
	find . -iname \*.pb.go \
	    -o -iname \*.pb.gw.go \
		-o -iname \*.mockgen.go \
		-o -iname \*.swagger.json \
		-exec rm -rf {} \;
	buf generate
	go generate ./...

.PHONY: setup
setup:
	go mod download
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %
	npm ci

.PHONY: docker-build
docker-build:
	docker build --platform linux/amd64 -f cmd/frontend/Dockerfile -t $(REGISTRY)/indexer/frontend .
	docker build --platform linux/amd64 -f cmd/worker/Dockerfile -t $(REGISTRY)/indexer/worker .

.PHONY: docker-push
docker-push:
	docker push $(REGISTRY)/indexer/frontend
	docker push $(REGISTRY)/indexer/worker

.PHONY: db-up
db-up:
	migrate -source $(MIGRATION_SOURCE) -database spanner://$(SPANNER_DATABASE)?x-clean-statements=true up

.PHONY: db-down-one
db-down-one:
	migrate -source $(MIGRATION_SOURCE) -database spanner://$(SPANNER_DATABASE)?x-clean-statements=true down 1

.PHONY: db-drop
db-drop:
	migrate -source $(MIGRATION_SOURCE) -database spanner://$(SPANNER_DATABASE)?x-clean-statements=true drop
