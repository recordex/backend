cd api/proto/pkg
git clone git@github.com:bufbuild/protoc-gen-validate.git
cd protoc-gen-validate && make build
cd ../../../..
protoc --proto_path=api/proto \
       --proto_path=api/proto/pkg/protoc-gen-validate \
       --go_out=api/gen --go_opt=paths=source_relative \
       --go-grpc_out=api/gen --go-grpc_opt=paths=source_relative \
       --validate_out="lang=go,paths=source_relative:api/gen" \
       api/proto/*.proto
