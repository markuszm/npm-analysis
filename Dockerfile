FROM golang:1.10.2 as builder
WORKDIR /go/src/github.com/markuszm/npm-analysis/
COPY . .
ARG exec_name
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o app ./${exec_name}

FROM scratch
COPY --from=builder /go/src/github.com/markuszm/npm-analysis/app .
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
ENTRYPOINT ["/app"]