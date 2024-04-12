build:
	cd container-base && make build
	cd dependencies && make build
	cd daemon && make build
	