build:
	sudo docker run --rm -v `pwd`:/workspace golang:1.21 sh -c "cd /workspace && go env -w GOPROXY=https://goproxy.cn,direct && go mod tidy && go build"
wrap:
	-sudo docker rmi satellite_emulator/master-node
	sudo docker build . -t satellite_emulator/master-node
run:
	sudo docker run \
		--rm \
		--privileged=true \
		-p 8080:8080 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-e MASTER_ADDRESS=10.134.148.56 \
		satellite_emulator/master-node