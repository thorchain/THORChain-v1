PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_TAGS = netgo ledger
BUILD_FLAGS = -tags "${BUILD_TAGS}" -ldflags "-X github.com/thorchain/THORChain/version.GitCommit=${COMMIT_HASH}"
GCC := $(shell command -v gcc 2> /dev/null)
LEDGER_ENABLED ?= true

########################################
### All

all: get_tools get_vendor_deps install test_lint test

########################################
### CI

ci: get_tools get_vendor_deps install test_cover test_lint test

########################################
### Build/Install

check-ledger:
ifeq ($(LEDGER_ENABLED),true)
ifndef GCC
$(error "gcc not installed for ledger support, please install")
endif
else
TMP_BUILD_TAGS := $(BUILD_TAGS)
BUILD_TAGS = $(filter-out ledger, $(TMP_BUILD_TAGS))
endif

build: check-ledger
ifeq ($(OS),Windows_NT)
	go build $(BUILD_FLAGS) -o build/thorchaind.exe ./cmd/thorchaind
	go build $(BUILD_FLAGS) -o build/thorchaincli.exe ./cmd/thorchaincli
else
	go build $(BUILD_FLAGS) -o build/thorchaind ./cmd/thorchaind
	go build $(BUILD_FLAGS) -o build/thorchaincli ./cmd/thorchaincli
endif

build-debug: check-ledger
ifeq ($(OS),Windows_NT)
	go build $(BUILD_FLAGS) -o build/thorchaindebug.exe ./cmd/thorchaindebug
else
	go build $(BUILD_FLAGS) -o build/thorchaindebug ./cmd/thorchaindebug
endif

build-spam: check-ledger
ifeq ($(OS),Windows_NT)
	go build $(BUILD_FLAGS) -o build/thorchainspam.exe ./cmd/thorchainspam
else
	go build $(BUILD_FLAGS) -o build/thorchainspam ./cmd/thorchainspam
endif

build-linux:
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

build-debug-linux:
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build-debug

build-spam-linux:
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build-spam

install: check-ledger
	go install $(BUILD_FLAGS) ./cmd/thorchaind
	go install $(BUILD_FLAGS) ./cmd/thorchaincli

install_debug:
	go install $(BUILD_FLAGS) ./cmd/thorchaindebug

install_spam:
	go install $(BUILD_FLAGS) ./cmd/thorchainspam

recreate: check-ledger
	rm -rf ~/.thorchaind
	make install
	thorchaind init --name local_validator --chain-id test-chain-local
	thorchaind start

########################################
### Tools & dependencies

check_tools:
	cd tools && $(MAKE) check_tools

check_dev_tools:
	cd tools && $(MAKE) check_dev_tools

update_tools:
	cd tools && $(MAKE) update_tools

update_dev_tools:
	cd tools && $(MAKE) update_dev_tools

get_tools:
	cd tools && $(MAKE) get_tools

get_dev_tools:
	cd tools && $(MAKE) get_dev_tools

get_vendor_deps:
	@echo "--> Running dep ensure"
	@dep ensure -v

draw_deps:
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i github.com/thorchain/THORChain/cmd/thorchaind -d 2 | dot -Tpng -o dependency-graph.png


########################################
### Documentation

godocs:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/thorchain/THORChain/app"
	godoc -http=:6060


########################################
### Testing

test: test_unit

test_cli:
	# TODO -vet=off is a temporary workaround to allow tests to pass until vendor/github.com/tendermint/iavl/proof_range.go is upgraded
	@go test -vet=off -count 1 -p 1 `go list github.com/thorchain/THORChain/cli_test` -tags=cli_test

test_unit:
	# TODO -vet=off is a temporary workaround to allow tests to pass until vendor/github.com/tendermint/iavl/proof_range.go is upgraded
	@go test -vet=off $(PACKAGES_NOSIMULATION)

test_race:
	# TODO -vet=off is a temporary workaround to allow tests to pass until vendor/github.com/tendermint/iavl/proof_range.go is upgraded
	@go test -vet=off -race $(PACKAGES_NOSIMULATION)

test_cover:
	@bash tests/test_cover.sh

test_lint:
	gometalinter.v2 --config=tools/gometalinter.json ./...
	!(gometalinter.v2 --disable-all --enable='errcheck' --vendor ./... | grep -v "client/")
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	dep status >> /dev/null

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs misspell -w

benchmark:
	# TODO -vet=off is a temporary workaround to allow tests to pass until vendor/github.com/tendermint/iavl/proof_range.go is upgraded
	@go test -vet=off -bench=. $(PACKAGES_NOSIMULATION)

########################################
### Local validator nodes using docker and docker-compose

build-docker-thorchaind-node:
	$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: localnet-stop
	@if ! [ -f build/node0/thorchaind/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/thorchaind:Z thorchain/thorchaindnode testnet --v 4 --o . --starting-ip-address 192.168.10.2 ; fi
	docker-compose -f ./networks/local/docker-compose.yml up -d

# Stop testnet
localnet-stop:
	docker-compose -f ./networks/local/docker-compose.yml down

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: all ci check-ledger build build-debug build-spam build-linux build-debug-linux build-spam-linux \
install install_debug install_spam recreate dist \
check_tools check_dev_tools update_tools update_dev_tools get_tools get_dev_tools get_vendor_deps draw_deps godocs \
test test_cli test_unit test_race test_cover test_lint format benchmark \
build-docker-thorchaind-node localnet-start localnet-stop
