all:
	bash build_tarball
	go get github.com/tools/godep
	godep restore
	cd cmd/enmasse-rest-server && go build

install:
	cd cmd/enmasse-rest-server && go install
