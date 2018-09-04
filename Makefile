PACKAGES=$(shell go list ./... | grep -v '/vendor/')
PACKAGES_NOCLITEST=$(shell go list ./... | grep -v '/vendor/' | grep -v github.com/thorchain/THORChain/cli_test)
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

build-spam: check-ledger
ifeq ($(OS),Windows_NT)
	go build $(BUILD_FLAGS) -o build/thorchainspam.exe ./cmd/thorchainspam
else
	go build $(BUILD_FLAGS) -o build/thorchainspam ./cmd/thorchainspam
endif

build-spam-linux:
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build-spam

build-linux:
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

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

update_tools:
	cd tools && $(MAKE) update_tools

get_tools:
	cd tools && $(MAKE) get_tools

get_vendor_deps:
	@rm -rf vendor/
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
	@go test -count 1 -p 1 `go list github.com/thorchain/THORChain/cli_test`

test_unit:
	@go test $(PACKAGES_NOCLITEST)

test_race:
	@go test -race $(PACKAGES_NOCLITEST)

test_cover:
	@bash tests/test_cover.sh

test_lint:
	gometalinter.v2 --config=tools/gometalinter.json ./...
	!(gometalinter.v2 --disable-all --enable='errcheck' --vendor ./... | grep -v "client/")
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs misspell -w

benchmark:
	@go test -bench=. $(PACKAGES_NOCLITEST)

########################################
### Local validator nodes using docker and docker-compose

build-docker-thorchaind-node:
	$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: localnet-stop
	@if ! [ -f build/node0/thorchaind/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/thorchaind:Z thorchain/thorchaindnode testnet --v 4 --o . --starting-ip-address 192.168.10.2 ; fi
	docker-compose -f ./networks/local/docker-compose.yml up

# Stop testnet
localnet-stop:
	docker-compose -f ./networks/local/docker-compose.yml down

########################################
### Remote validator nodes using terraform and ansible

TESTNET_NAME?=remotetestnet
SERVERS?=4
SSH_KEY_NAME?="$(TESTNET_NAME)-deployer"
SSH_PRIVATE_FILE?="$(HOME)/.ssh/id_rsa"
SSH_PUBLIC_FILE?="$(HOME)/.ssh/id_rsa.pub"
BINARY=$(CURDIR)/build/thorchaind

remotenet-start:
	@if [ -z "$(AWS_SECRET_KEY)" ]; then echo "AWS_SECRET_KEY environment variable not set." ; false ; fi
	@if [ -z "$(AWS_ACCESS_KEY)" ]; then echo "AWS_ACCESS_KEY environment variable not set." ; false ; fi
	@if ! [ -f $(SSH_PUBLIC_FILE) ]; then ssh-keygen ; fi
	@if [ -z "`file $(BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-linux'." ; false ; fi
	cd networks/remote/terraform && terraform init && terraform apply -var TESTNET_NAME="$(TESTNET_NAME)" -var SERVERS="$(SERVERS)" -var AWS_SECRET_KEY="$(AWS_SECRET_KEY)" -var AWS_ACCESS_KEY="$(AWS_ACCESS_KEY)" -var SSH_KEY_NAME="$(SSH_KEY_NAME)" -var SSH_PRIVATE_FILE="$(SSH_PRIVATE_FILE)" -var SSH_PUBLIC_FILE="$(SSH_PUBLIC_FILE)"
	cd networks/remote/terraform && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i /usr/local/bin/terraform-inventory -e BINARY=$(BINARY) -e TESTNET_NAME="$(TESTNET_NAME)" ../ansible/setup-validators.yml
	cd networks/remote/terraform && ansible-playbook -i /usr/local/bin/terraform-inventory ../ansible/set-toml-values.yml
	cd networks/remote/terraform && ansible-playbook -i /usr/local/bin/terraform-inventory -e TESTNET_NAME="$(TESTNET_NAME)" ../ansible/start.yml

remotenet-reset-with-genesis:
	@if ! [ -f $(GENESIS_FILE) ]; then echo "GENESIS environment variable not set." ; false ; fi
	cd networks/remote/terraform && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i /usr/local/bin/terraform-inventory -e GENESIS_FILE=$(GENESIS_FILE) ../ansible/reset-validators-with-genesis.yml
	cd networks/remote/terraform && ansible-playbook -i /usr/local/bin/terraform-inventory ../ansible/set-toml-values.yml

remotenet-stop:
	@if [ -z "$(AWS_SECRET_KEY)" ]; then echo "AWS_SECRET_KEY environment variable not set." ; false ; fi
	@if [ -z "$(AWS_ACCESS_KEY)" ]; then echo "AWS_ACCESS_KEY environment variable not set." ; false ; fi
	cd networks/remote/terraform && terraform destroy -var AWS_SECRET_KEY="$(AWS_SECRET_KEY)" -var AWS_ACCESS_KEY="$(AWS_ACCESS_KEY)" -var SSH_KEY_NAME="$(SSH_KEY_NAME)" -var SSH_PRIVATE_FILE="$(SSH_PRIVATE_FILE)" -var SSH_PUBLIC_FILE="$(SSH_PUBLIC_FILE)"

remotenet-status:
	cd networks/remote/terraform && ansible-playbook -i /usr/local/bin/terraform-inventory ../ansible/status.yml

########################################
### Remote spam nodes using terraform and ansible

SPAM_CLUSTER_NAME?="$(TESTNET_NAME)-spam"
SERVERS?=4
SPAM_SSH_KEY_NAME?="$(SPAM_CLUSTER_NAME)-deployer"
SSH_PRIVATE_FILE?="$(HOME)/.ssh/id_rsa"
SSH_PUBLIC_FILE?="$(HOME)/.ssh/id_rsa.pub"
SPAM_BINARY=$(CURDIR)/build/thorchainspam
CLI_BINARY=$(CURDIR)/build/thorchaincli

remotenet-spam-start:
	@if [ -z "$(AWS_SECRET_KEY)" ]; then echo "AWS_SECRET_KEY environment variable not set." ; false ; fi
	@if [ -z "$(AWS_ACCESS_KEY)" ]; then echo "AWS_ACCESS_KEY environment variable not set." ; false ; fi
	@if ! [ -f $(SSH_PUBLIC_FILE) ]; then ssh-keygen ; fi
	@if [ -z "`file $(SPAM_BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-spam-linux'." ; false ; fi
	@if [ -z "`file $(CLI_BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-linux'." ; false ; fi
	cd networks/remote-spam/terraform && terraform init && terraform apply -var CLUSTER_NAME="$(SPAM_CLUSTER_NAME)" -var SERVERS="$(SERVERS)" -var AWS_SECRET_KEY="$(AWS_SECRET_KEY)" -var AWS_ACCESS_KEY="$(AWS_ACCESS_KEY)" -var SSH_KEY_NAME="$(SPAM_SSH_KEY_NAME)" -var SSH_PRIVATE_FILE="$(SSH_PRIVATE_FILE)" -var SSH_PUBLIC_FILE="$(SSH_PUBLIC_FILE)"
	cd networks/remote-spam/terraform && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i /usr/local/bin/terraform-inventory -e SPAM_BINARY=$(SPAM_BINARY) -e CLI_BINARY=$(CLI_BINARY) -e CLUSTER_NAME="$(SPAM_CLUSTER_NAME)" ../ansible/setup-spammers.yml

remotenet-spam-stop:
	@if [ -z "$(AWS_SECRET_KEY)" ]; then echo "AWS_SECRET_KEY environment variable not set." ; false ; fi
	@if [ -z "$(AWS_ACCESS_KEY)" ]; then echo "AWS_ACCESS_KEY environment variable not set." ; false ; fi
	cd networks/remote-spam/terraform && terraform destroy -var AWS_SECRET_KEY="$(AWS_SECRET_KEY)" -var AWS_ACCESS_KEY="$(AWS_ACCESS_KEY)" -var SSH_KEY_NAME="$(SSH_KEY_NAME)" -var SSH_PRIVATE_FILE="$(SSH_PRIVATE_FILE)" -var SSH_PUBLIC_FILE="$(SSH_PUBLIC_FILE)"

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: build install install_debug dist \
check_tools get_tools get_vendor_deps draw_deps test test_cli test_unit \
test_cover test_lint benchmark \
build-linux build-docker-thorchaindnode localnet-start localnet-stop remotenet-start \
remotenet-stop remotenet-status format check-ledger
