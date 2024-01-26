build:
	cd ./daemon && go build
	-mkdir scripts
	mv ./daemon/satellite ./scripts/entrypoint
	sudo docker build --no-cache . -t realssd/satellite-router
clean:
	rm ./scripts/entrypoint
