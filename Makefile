all: build

proton:
	bash build_tarball

build: proton
	go get github.com/tools/godep
	godep restore
	cd cmd/enmasse-rest-server && go build

install:
	cd cmd/enmasse-rest-server && go install


test:
	cd tests/unit && go test
