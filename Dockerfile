FROM golang:1.10

ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

WORKDIR $GOPATH/src/github.com/baystation12/byond-exporter
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . ./
RUN GOOS=linux go build -a -o byond-exporter .

ENTRYPOINT ["./byond-exporter"]
