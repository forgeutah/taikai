#!/usr/bin/env bash

set -e

THISSCRIPT=$(basename $0)

GITHUB_OUTPUT=${GITHUB_OUTPUT:-$(mktemp)}

# Modify for the help message
usage() {
  echo "${THISSCRIPT} command"
  echo "Executes the step command in the script."
  exit 0
}

install_proto_tools() {
  echo "Installing proto tooling"
  go install github.com/catalystsquad/protoc-gen-go-gorm@latest
	go install github.com/favadi/protoc-go-inject-tag@latest
	go install github.com/mitchellh/protoc-gen-go-json@latest
	npm install @bufbuild/protobuf @bufbuild/protoc-gen-es @bufbuild/buf @openapitools/openapi-generator-cli
}

build_protos() {
  install_proto_tools
  echo "Building proto generated things"
  PATH="$PATH:$(pwd)/node_modules/.bin"
	cd protos; buf generate;protoc-go-inject-tag -input "gen/go/taikai/v1/*.*.*.go";protoc-go-inject-tag -input "gen/go/taikai/v1/*.*.go"; cd -
	#mkdir -p protos/clients/js; cd protos/clients/js; openapi-generator-cli generate -g javascript -i ../../gen/docs/taikai/v1/api.swagger.json -c config.yaml; npm install; npm run build; cd -
	# mkdir -p protos/clients/ts; cd protos/clients/ts; openapi-generator-cli generate -g typescript-fetch -i ../../gen/docs/taikai/v1/api.swagger.json -c config.yaml; npm install; npm run build; cd -
}

run() {
 go mod tidy; cd server; go run main.go; cd -
}

test() {
	go test -v -cover ./...
}

setup() {
  build_protos
  cd helm_chart; helm dependency build ./; cd -
  go mod tidy;
}

golint() {
	echo "Linting"
	go mod tidy
	golangci-lint run
}

buflint() {
	echo "Linting"
	cd protos; buf lint
}

# This should be last in the script, all other functions are named beforehand.
case "$1" in
  "install_proto_tools")
    shift
    install_proto_tools "$@"
    ;;
  "build_protos")
    shift
    build_protos "$@"
    ;;
	"buflint")
		shift
		buflint "$@"
		;;
	"golint")
		shift
		golint "$@"
		;;
  "run")
    shift
    run "$@"
    ;;
 "setup")
    shift
    setup "$@"
    ;;
	"test")
		shift
		test "$@"
		;;
  *)
    usage
    ;;
esac

exit 0
