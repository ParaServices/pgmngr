# builder
FROM golang:1.13.0-alpine3.10 as builder

LABEL maintainer="kareem@joinpara.com"

RUN apk add --no-cache git=2.22.0-r0

RUN mkdir -p /go/src/github.com/ParaServices/pgmngr/

WORKDIR /go/src/github.com/ParaServices/pgmngr/

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod vendor -v -a -installsuffix cgo -o pgmngr .

# actual container
FROM alpine:3.10

RUN apk add --no-cache bash=5.0.0-r0

RUN mkdir -p /pgmngr

WORKDIR /pgmngr

COPY --from=builder /go/src/github.com/ParaServices/pgmngr/ .
ENV PATH="${PATH}:/pgmngr"
