build:
	sudo docker run --rm -v `pwd`:/workspace golang:1.20 sh -c "cd /workspace && go env -w GOPROXY=https://goproxy.cn,direct && go mod tidy && go build"
wrap:
	-sudo docker rmi -f satellite_emulator/instance-manager
	sudo docker build . -t satellite_emulator/instance-manager
run:
	sudo docker run \
		--rm \
		--privileged=true \
		-v /var/run/docker.sock:/var/run/docker.sock \
		satellite_emulator/instance-manager