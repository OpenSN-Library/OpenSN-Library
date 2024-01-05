build:
	cd ./daemon && go build
	mv ./daemon/satellite ./scripts/entrypoint
	sudo docker build . -t `basename \`pwd\``
clean:
	rm ./scripts/entrypoint