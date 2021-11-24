FROM golang:1.17 as builder
WORKDIR /go/src/github.com/agravelot/permis_toolkit
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app cmd/bot/main.go

FROM chromedp/headless-shell
RUN apt update && apt install -y ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/agravelot/permis_toolkit .
ENTRYPOINT ["./app"]   