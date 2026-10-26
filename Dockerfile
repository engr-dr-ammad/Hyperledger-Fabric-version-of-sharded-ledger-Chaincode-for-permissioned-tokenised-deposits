FROM golang:1.20-alpine AS builder

WORKDIR /opt/gopath/src/deposit-chaincode
COPY chaincode .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o deposit-chaincode .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /opt/gopath/src/deposit-chaincode/deposit-chaincode .

EXPOSE 7052

CMD ["./deposit-chaincode"]
