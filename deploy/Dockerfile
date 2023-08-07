FROM golang:1.20 as builder
WORKDIR /go/src/app
COPY . .

# TODO: enable for production
#RUN go test -v ./...
RUN go get -d -v ./...

# build with version
ARG GIT_TAG
ARG BUILD_TIME
RUN CGO_ENABLED=0 go build -ldflags="-X main.version=$GIT_TAG -X main.buildTime=$BUILD_TIME" -a -installsuffix cgo -o app .
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/app .
CMD ["./app"]