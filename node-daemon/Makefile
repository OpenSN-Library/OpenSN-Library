build:
	mkdir -p build/config
	echo "Build Commit:"`git rev-parse HEAD`", Build Time:"`date`", Arch:"`go env GOARCH`"-"`go env GOOS` > share/static/dist/version.txt
	go build
	mv ./NodeDaemon ./build/NodeDaemon
	cp ./config/config.json ./build/config/
	cp share/static/dist/version.txt ./build/
	tar cvf satellite-emulator-daemon.tar.gz build/*
clean:
	-sudo rm -rf ./build
	-sudo rm satellite-emulator-daemon.tar.gz
wrap:
	sudo docker run --rm -v `pwd`:/workspace golang:1.21 sh -c "cd /workspace && go env -w GOPROXY=https://goproxy.cn,direct && go mod tidy && go build"
	-sudo docker rmi -f satellite_emulator/node-daemon
	sudo docker build . -t satellite_emulator/node-daemon
run:
	cd build && sudo ./NodeDaemon
dep:
	-sudo docker rm -f emulator_etcd
	-sudo docker rm -f  emulator_redis
	sudo docker run -d --rm --name=emulator_etcd --cpus=0.25 -m 128m -p 2379:2379 satellite_emulator/etcd
	sudo docker run -d --rm --name=emulator_redis --cpus=0.25 -m 128m -p 6379:6379 satellite_emulator/redis
