# Copyright © 2021 Kris Nóva <kris@nivenly.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#  ────────────────────────────────────────────────────────────────────────────
#
#   ████████╗██╗    ██╗██╗███╗   ██╗██╗  ██╗
#   ╚══██╔══╝██║    ██║██║████╗  ██║╚██╗██╔╝
#      ██║   ██║ █╗ ██║██║██╔██╗ ██║ ╚███╔╝
#      ██║   ██║███╗██║██║██║╚██╗██║ ██╔██╗
#      ██║   ╚███╔███╔╝██║██║ ╚████║██╔╝ ██╗
#      ╚═╝    ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝
#
#  ────────────────────────────────────────────────────────────────────────────


all: compile
version=$(shell git rev-parse HEAD)

# Global release version.
# Change this to bump the build version!
version="0.3.1"

compile: generate ## Compile for the local architecture ⚙
	@echo "Compiling..."
	go build -ldflags "-X 'github.com/kris-nova/twinx.Version=$(version)'" -o twinx cmd/*.go

install: ## Install your twinx 🎉
	@echo "Installing..."
	cp twinx /usr/local/bin/twinx

generate: ## Will generate Go code from .proto files in /api
	@echo "Generating..."
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
        --go-grpc_out=. \
        --go-grpc_opt=paths=source_relative \
        api/twinx.proto


test: ## 🤓 Test is used to test
	@echo "Testing..."
	go test -v ./...

clean: ## Clean your artifacts 🧼
	@echo "Cleaning..."
	rm -rf release
	rm -rf embed_*
	rm -rf twinx
	rm -rf app
	rm -rf out/*

release: ## Make the binaries for a GitHub release 📦
	mkdir -p release
	GOOS="linux" GOARCH="amd64" go build -ldflags "-X 'github.com/kris-nova/twinx.Version=$(version)'" -o release/twinx-linux-amd64 cmd/*.go
	GOOS="linux" GOARCH="arm" go build -ldflags "-X 'github.com/kris-nova/twinx.Version=$(version)'" -o release/twinx-linux-arm cmd/*.go
	GOOS="linux" GOARCH="arm64" go build -ldflags "-X 'github.com/kris-nova/twinx.Version=$(version)'" -o release/twinx-linux-arm64 cmd/*.go
	GOOS="linux" GOARCH="386" go build -ldflags "-X 'github.com/kris-nova/twinx.Version=$(version)'" -o release/twinx-linux-386 cmd/*.go
	GOOS="darwin" GOARCH="amd64" go build -ldflags "-X 'github.com/kris-nova/twinx.Version=$(version)'" -o release/twinx-darwin-amd64 cmd/*.go


.PHONY: help
help:  ## 🤔 Show help messages for make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}'
