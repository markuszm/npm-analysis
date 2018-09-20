FROM golang:1 as builder

WORKDIR /npm-analysis

# Caching dependencies
COPY go.mod .
COPY go.sum .
RUN go get -v ./...

# Building binary
COPY . .
ARG exec_name
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -o app ./${exec_name}

# Copying binary to scratch image
FROM scratch
COPY --from=builder /npm-analysis/app .
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
ENTRYPOINT ["/app"]