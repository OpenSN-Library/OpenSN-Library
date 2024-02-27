build:
	cd ./daemon && go build
	-mkdir scripts
	mv ./daemon/ground ./scripts/entrypoint
	sudo docker build --no-cache . -t realssd/ground-station
clean:
	rm ./scripts/entrypoint
