.PHONY: init
init:
	brew install bufbuild/buf/buf
	go mod vendor
	go install github.com/sudorandom/protoc-gen-connect-openapi@main

.PHONY: proto
proto:
	buf generate
	buf format --write
	npm run generate-barrels

protoc:
	PATH=$PATH:$(pwd)/node_modules/.bin \
      protoc -I . \
      --es_out src/gen \
      --es_opt target=ts \
      --connect-es_out src/gen \
      --connect-es_opt target=ts \
      proto/playground/v1/*.proto