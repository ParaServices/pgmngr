# builder
FROM golang:1.15.2-alpine3.12 as builder

LABEL maintainer="kareem@joinpara.com"

# remove testdata folder, we only need this for dev
RUN rm -rf testdata

RUN apk add --no-cache git

RUN mkdir -p /go/src/github.com/ParaServices/pgmngr/

WORKDIR /go/src/github.com/ParaServices/pgmngr/

COPY . .

RUN mkdir bin

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod vendor -v -a -installsuffix cgo -o ./bin/pgmngr

# actual container
FROM alpine:3.12

RUN apk add --no-cache bash

RUN mkdir -p /pgmngr

WORKDIR /pgmngr

COPY --from=builder /go/src/github.com/ParaServices/pgmngr/bin/pgmngr .
ENV PATH="${PATH}:/pgmngr"
