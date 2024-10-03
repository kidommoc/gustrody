FROM golang:1.21.6-alpine

# China specified
RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN echo "#!/bin/sh" >> /loop.sh
RUN echo "go mod download"
RUN echo "while true; do sleep 10; done" >> /loop.sh
RUN chmod a+x /loop.sh
WORKDIR /app
VOLUME [ "/app" ]
ENTRYPOINT ["/loop.sh"]