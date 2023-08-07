FROM golang:1.20 as builder
WORKDIR /go/src/app
COPY . .

RUN go test -v ./...
RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/app .
CMD ["./app"]