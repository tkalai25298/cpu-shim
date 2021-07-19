# External build of chaincode

ARG GO_VER=1.15.2
ARG ALPINE_VER=3.12

FROM golang:${GO_VER}-alpine${ALPINE_VER}

WORKDIR /go/src/github.com/sachin-ngpws/cpu-shim
COPY . .

RUN go build -mod vendor -o cpu
RUN mv cpu /go/bin/
RUN ls -al /go/bin

EXPOSE 7054
CMD ["cpu"]