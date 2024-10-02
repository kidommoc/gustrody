#!/bin/sh
docker volume create austrody_main
docker volume create austrody_redis
docker compose up -d