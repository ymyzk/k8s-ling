FROM golang:1.14-alpine as build

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY ./go.mod ./go.sum ./*.go ./*.html /go/src/github.com/ymyzk/k8s-ling/
WORKDIR /go/src/github.com/ymyzk/k8s-ling

RUN go build -o /bin/k8s-ling

FROM scratch

#COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/k8s-ling /app/k8s-ling
COPY --from=build /go/src/github.com/ymyzk/k8s-ling/*.html /app/

# A bit dirty hack: setting WORKDIR for discovering templates
WORKDIR /app
ENTRYPOINT ["/app/k8s-ling"]
