FROM golang:1.13

ENV GO111MODULE=on

WORKDIR /exporter
COPY . ./
RUN GOOS=linux go build -a -o byond-exporter .

ENTRYPOINT ["./byond-exporter"]
