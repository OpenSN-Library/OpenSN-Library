build:
	go build
wrap:
	sudo docker run --rm -v `pwd`:/workspace golang:1.21 sh -c "cd /workspace && go env -w GOPROXY=https://goproxy.cn,direct && go mod tidy && go build"
	-sudo docker rmi -f satellite_emulator/node-daemon
	sudo docker build . -t satellite_emulator/node-daemon
run:
	sudo ./NodeDaemon
dep:
	-sudo docker stop emulator_etcd
	-sudo docker stop emulator_redis
	sudo docker run -d --rm --name=emulator_etcd --cpus=0.25 -m 128m -p 2379:2379 satellite_emulator/etcd
	sudo docker run -d --rm --name=emulator_redis --cpus=0.25 -m 128m -p 6379:6379 satellite_emulator/redis
