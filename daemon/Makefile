build:
	mkdir -p opensn-daemon/config
	echo "Build Commit:"`git rev-parse HEAD`", Build Time:"`date`", Arch:"`go env GOARCH`"-"`go env GOOS` > share/static/dist/version.txt
	go build
	mv ./NodeDaemon ./opensn-daemon/NodeDaemon
	cp ./config/config.json ./opensn-daemon/config/
	cp share/static/dist/version.txt ./opensn-daemon/
	cp depend-bin/gotty ./opensn-daemon/
	cp -r scripts ./opensn-daemon/
	tar cvf opensn-daemon.tar.gz opensn-daemon/*
clean:
	-sudo rm -rf ./opensn-daemon
	-sudo rm opensn-daemon.tar.gz
dep:
	-sudo docker rm -f opensn_etcd
	-sudo docker rm -f opensn_influxdb
	-sudo docker rm -f opensn_codeserver
	sudo docker run -d --rm --name=opensn_etcd -p 2379:2379 realssd/opensn_etcd
	sudo docker run -d --rm --name=opensn_influxdb  -p 8086:8086 realssd/opensn_influxdb
	sudo docker run -d --rm --name=opensn_codeserver -v `pwd`/../TopoConfigurators:/workspace --cpus=1 -m 1g -p 8079:8443 realssd/opensn_codeserver
