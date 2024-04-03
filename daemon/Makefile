build:
	mkdir -p build/config
	echo "Build Commit:"`git rev-parse HEAD`", Build Time:"`date`", Arch:"`go env GOARCH`"-"`go env GOOS` > share/static/dist/version.txt
	go build
	mv ./NodeDaemon ./build/NodeDaemon
	cp ./config/config.json ./build/config/
	cp share/static/dist/version.txt ./build/
	cp depend-bin/gotty ./build/
	tar cvf satellite-emulator-daemon.tar.gz build/*
clean:
	-sudo rm -rf ./build
	-sudo rm satellite-emulator-daemon.tar.gz
wrap: build
	-sudo docker rmi -f satellite_emulator/node-daemon
	sudo docker build . -t satellite_emulator/node-daemon
run:
	cd build && sudo ./NodeDaemon
run-docker:
	-sudo docker rm -f node-daemon
	sudo docker run \
		--net=host \
		-p 8080:8080 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /home/satellite/Workspace/satellite-emulator/node-daemon/build/runtime:/runtime/runtime \
		--privileged=true \
		--name=node-daemon \
		satellite_emulator/node-daemon
run-sidecar:
	cd ../sub-modules/PositionProvider/ && make run
dep:
	-sudo docker rm -f opensn_etcd
	-sudo docker rm -f opensn_influxdb
	-sudo docker rm -f opensn_codeserver
	sudo docker run -d --rm --name=opensn_etcd -p 2379:2379 realssd/opensn_etcd
	sudo docker run -d --rm --name=opensn_influxdb --cpus=0.5 -m 256m -p 8086:8086 realssd/opensn_influxdb
	sudo docker run -d --rm --name=opensn_codeserver --cpus=1 -m 1g -p 8079:8443 realssd/opensn_codeserver
