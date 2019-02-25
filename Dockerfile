FROM golang:1.11 as resource
WORKDIR /go/src/github.com/dpb587/metalink-repository-resource
COPY . .
ENV CGO_ENABLED=0
RUN mkdir -p /opt/resource
RUN git rev-parse HEAD | tee /opt/resource/version
RUN go build -o /opt/resource/check ./check
RUN go build -o /opt/resource/in ./in
RUN go build -o /opt/resource/out ./out

FROM alpine:3.4
RUN apk --no-cache add bash ca-certificates git openssh-client
COPY --from=resource /opt/resource /opt/resource
RUN mkdir ~/.ssh && echo "StrictHostKeyChecking no" > ~/.ssh/config
