FROM redis:7.2.4-alpine

EXPOSE 6739
ARG DB
ENV CONF_DIR=/usr/local/etc
ENV SECRET_PATH=/run/secrets/redis

WORKDIR /tmp
COPY ./redis-${DB}.conf ./redis.conf
RUN --mount=type=secret,id=redis \
    REDIS_SECRET=$(cat /run/secrets/redis) && \
    sed -e "s/{REDIS_SECRET}/$REDIS_SECRET/g" < ./redis.conf > ./new.conf
RUN mkdir -p ${CONF_DIR} && cp ./new.conf ${CONF_DIR}/redis.conf

CMD ["redis-server", "/usr/local/etc/redis.conf"]