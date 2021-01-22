TKX = go run ./cmd/tkx -p ./examples

fmt:
	$(TKX) fmt -l -w ./examples

get:
	$(TKX) -v get github.com/rancher/local-path-provisioner@v0.0.18
	$(TKX) -v get ./...
	$(TKX) -v get -u ./...
	$(TKX) -v get ./...

show:
	$(TKX) show ./examples/clusters/demo/hello-world.jsonnet

export:
	$(TKX) show ./examples/clusters/demo/*.jsonnet -o ./_output

apply:
	$(TKX) apply ./examples/clusters/demo/hello-world.jsonnet

prune:
	$(TKX) prune ./examples/clusters/demo/hello-world.jsonnet

delete:
	$(TKX) delete ./examples/clusters/demo/hello-world.jsonnet

build:
	goreleaser build --snapshot --rm-dist

test:
	go test -v -race ./...

cover:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...

install: build
	mv ./bin/tankax_$(shell go env GOOS)_$(shell go env GOARCH)/tkx ${GOPATH}/bin/tkx

dep:
	go get -u ./...