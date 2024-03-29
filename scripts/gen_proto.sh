#!/bin/bash

set -e
mkdir -p api/proto/pkg
cd api/proto/pkg || exit
if [ ! -d "protoc-gen-validate" ]; then
  git clone git@github.com:bufbuild/protoc-gen-validate.git
fi
cd protoc-gen-validate && make build
cd ../../../..
protoc --proto_path=api/proto \
       --proto_path=api/proto/pkg/protoc-gen-validate \
       --go_out=api/gen --go_opt=paths=source_relative \
       --go-grpc_out=api/gen --go-grpc_opt=paths=source_relative \
       --validate_out="lang=go,paths=source_relative:api/gen" \
       api/proto/*.proto
