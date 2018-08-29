FROM golang:1 as builder
WORKDIR /npm-analysis
COPY . .
ARG exec_name
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o app ./${exec_name}

FROM scratch
COPY --from=builder /npm-analysis/app .
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
ENTRYPOINT ["/app"]