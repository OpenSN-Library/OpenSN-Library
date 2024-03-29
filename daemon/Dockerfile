FROM golang:1.20
COPY build /runtime/
EXPOSE 8080
WORKDIR /runtime
ENTRYPOINT [ "./NodeDaemon" ]