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