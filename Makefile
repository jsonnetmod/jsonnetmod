JMOD = go run ./cmd/jmod -v -p ./examples


json:
	$(JMOD) build ./examples/clusters/demo/hello-world.jsonnet

k.show:
	$(JMOD) k show ./examples/clusters/demo/hello-world.jsonnet

k.apply:
	$(JMOD) k apply ./examples/clusters/demo/hello-world.jsonnet

k.delete:
	$(JMOD) k delete ./examples/clusters/demo/hello-world.jsonnet

k.prune:
	$(JMOD) k prune ./examples/clusters/demo/hello-world.jsonnet

k.export:
	$(JMOD) k show -o ./_output ./examples/clusters/demo/*.jsonnet

fmt:
	$(JMOD) fmt -l -w .

build:
	goreleaser build --snapshot --rm-dist

test:
	go test -v -race ./pkg/...

cover:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./pkg/...

install: build
	mv ./bin/jsonnetmod_$(shell go env GOOS)_$(shell go env GOARCH)/jmod ${GOPATH}/bin/jmod

dep:
	go get -u ./...

fork: install.forkinternal
	forkinternal cmd/go/internal/modload cmd/go/internal/modfetch

install.forkinternal:
	go install github.com/jsonnetmod/tools/cmd/forkinternal@3e67f04

install.dev-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/go-courier/husky/cmd/husky@latest