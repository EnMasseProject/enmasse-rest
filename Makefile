all:
	./build_tarball
	godep restore
	cd cmd/enmasse-rest-server && go build

install:
	cd cmd/enmasse-rest-server && go install
