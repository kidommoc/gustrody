FROM golang:1.21.6-alpine

RUN echo "#!/bin/sh" >> /loop.sh
RUN echo "while true; do sleep 100; done" >> /loop.sh
RUN chmod a+x /loop.sh
WORKDIR /app
VOLUME [ "/app" ]
CMD ["/loop.sh"]