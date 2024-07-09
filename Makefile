build:
	mkdir opensn_build
	cd ui && make build
	cp -r ui/dist/ daemon/static/
	cd container-base && make build
	cd dependencies && make build
	cd daemon && make build
	mkdir -p opensn_build/node-images
	cp container-base/*.tar.gz opensn_build/node-images/
	cp -r daemon/opensn-daemon opensn_build/
	mkdir -p opensn_build/depend-images
	cp dependencies/*.tar.gz opensn_build/depend-images/
	cp -r TopoConfigurators opensn_build/
	cp -r tools opensn_build/
	tar cvf opensn_build.tar.gz opensn_build/*

clean:
	-rm -rf opensn_build
	-rm opensn_build.tar.gz
	cd ui && make clean
	cd container-base && make clean
	cd dependencies && make clean
	cd daemon && make clean