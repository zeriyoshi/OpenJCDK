FROM golang:1.25.3-trixie AS builder

WORKDIR /build

COPY go.mod go.sum bot.go ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot .

FROM scratch AS runner

COPY --from=builder /build/bot /bot
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/bot"]
