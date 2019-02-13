FROM golang:1.11.4-alpine

RUN apk add --no-cache gcc musl-dev git bash nodejs yarn curl

# Remove after upgrading nebula to master
RUN apk add --no-cache py-pip && pip install awscli

WORKDIR /go/src/github.com/orbs-network/performance-benchmark/galileo

ADD . /go/src/github.com/orbs-network/performance-benchmark/

RUN yarn install

CMD node index.js