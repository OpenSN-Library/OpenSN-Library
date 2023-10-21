build:
	sudo docker run --rm -v `pwd`:/workspace golang:1.20 sh -c "cd /workspace && go env -w GOPROXY=https://goproxy.cn,direct && go build"
wrap: build
	sudo ./build_dockerfile.sh