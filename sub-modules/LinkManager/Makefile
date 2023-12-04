bpf_src=bpf

bpf:
	clang -target bpf \
		-Wno-pointer-sign \
		-Wno-compare-distinct-pointer-types \
		-Werror \
		-O2 \
		-I /usr/src/linux-headers-6.2.0-35-generic/ \
		-emit-llvm \
		-c -g -o bpf/${bpf_src}.ll bpf-src/${bpf_src}.c 
	llc -march=bpf -filetype=obj -o bpf/${bpf_src}.o bpf/${bpf_src}.ll
	rm bpf/${bpf_src}.ll
clean:
	rm bpf/${bpf_src}.o

# veth1(test_0) <---veth peer---> veth0(init) <---tc redirect---> veth3(test_0) <---veth peer---> veth2(init)
# veth1(test_0) <---tc redirect---> veth3(init) <---veth peer---> veth2(init)
test_env:
	sudo ip netns add test_0
	sudo ip link add veth0 type veth peer name veth1
	sudo ip link add veth2 type veth peer name veth3
	sudo ip link set veth1 netns test_0
	sudo ip link set veth3 netns test_0
	sudo ip addr add 192.168.233.1/24 dev veth0
	sudo ip addr add 192.168.234.1/24 dev veth2
	sudo nsenter --net=/var/run/netns/test_0 ip addr add 192.168.233.2/24 dev veth1
	sudo nsenter --net=/var/run/netns/test_0 ip addr add 192.168.234.2/24 dev veth3
	sudo ip link set veth0 up
	sudo ip link set veth2 up
	sudo nsenter --net=/var/run/netns/test_0 ip link set veth1 up
	sudo nsenter --net=/var/run/netns/test_0 ip link set veth3 up

env_attach_ingress: all
	tc qdisc add dev veth1 clsact
	tc filter add dev veth1 ingress bpf da obj tc_veth_redirect.o sec ingress

env_detach_ingress:

env_test_ingress:
	sudo nsenter --net=/var/run/netns/test_0 ping 192.168.233.1

env_attach_egress: all
	tc qdisc add dev veth1 clsact
	tc filter add dev veth1 ingress bpf da obj tc_veth_redirect.o sec ingress

env_detach_egress:


env_test_egress:
	sudo nsenter --net=/var/run/test_0 ping 192.168.233.1

test_env_clean:
	sudo ip link del veth0
	sudo ip link del veth2
	sudo ip netns del test_0
build:
	sudo docker run --rm -v `pwd`:/workspace golang:1.21 sh -c "cd /workspace && go env -w GOPROXY=https://goproxy.cn,direct && go mod tidy && go build"
wrap:
	sudo docker build . -t satellite_emulator/link-manager
run:
	sudo docker run \
		--rm \
		--privileged=true \
		-v /var/run/docker.sock:/var/run/docker.sock \
		satellite_emulator/link-manager
# sudo nsenter --net=/var/run/netns/test_0 tc qdisc add dev veth3 clsact
# sudo nsenter --net=/var/run/netns/test_0 tc filter add dev veth3 egress bpf da obj tc_veth_redirect.o sec egress