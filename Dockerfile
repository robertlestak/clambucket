FROM golang:1.22 as builder

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /clambucket cmd/clambucket/*.go

FROM debian:bullseye-slim as app

RUN apt-get update && apt-get install -y ca-certificates clamav-daemon && rm -rf /var/lib/apt/lists/*

COPY --from=builder /clambucket /clambucket

CMD ["/clambucket"]