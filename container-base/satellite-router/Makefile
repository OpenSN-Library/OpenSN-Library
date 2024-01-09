build:
	cd ./daemon && go build
	mv ./daemon/satellite ./scripts/entrypoint
	sudo docker build --no-cache . -t realssd/satellite-router
clean:
	rm ./scripts/entrypoint
