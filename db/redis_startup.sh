#!/bin/sh
SECRET=$(cat $SECRET_FILE)
mkdir -p /usr/local/etc
sed -e "s/{REDIS_SECRET}/$SECRET/g" < /redis.conf > $CONFIG_FILE
docker-entrypoint.sh $CONFIG_FILE