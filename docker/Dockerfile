FROM alpine:3.4
RUN apk --no-cache add bash ca-certificates git openssh-client
ADD assets/ /opt/resource
RUN mkdir ~/.ssh && echo "StrictHostKeyChecking no" > ~/.ssh/config
