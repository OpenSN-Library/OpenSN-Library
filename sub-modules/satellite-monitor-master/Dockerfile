FROM golang:1.20

COPY . /monitor
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN cd /monitor && go mod tidy
EXPOSE 8080:8080
WORKDIR /monitor
CMD ./monitor 30000 8080 > log.log