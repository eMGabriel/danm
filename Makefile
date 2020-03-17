
CRDDIR=github.com/nokia/danm/pkg/crd
APISDIR=$(CRDDIR)/apis
DANMV1DIR=$(CRDDIR)/apis/danm/v1
DEEPCOPYGENDILE=zz_generated.deepcopy
CLIENTDIR=$(CRDDIR)/client


all: danm webhook netwatcher svcwatcher

.PHONY: code-gen
code-gen:
	bash hack/update-codegen.sh

.PHONY: test
test: coverage.out verify-code-gen

.PHONY: verify-code-gen
verify-code-gen:
	bash hack/verify-codegen.sh

coverage.out:
	go test -v -cover -covermode=count -coverprofile=coverage.out  -coverpkg=./pkg/... ./test/uts/...
	#go test -v -coverpkg ./... ./...

.PHONY: clean
clean:
	rm -rf coverage.out vendor

.PHONY: bin
bin:
	rm ./bin
	docker build --target=binaries --tag=binaries ./scm/build
	docker run --rm -v ${PWD}:/mnt/project binaries cp /go/bin/* /mnt/project/
