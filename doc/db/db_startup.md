# Database Startup

Temporary document before `docker compose` applied.

## Create docker network

```sh
$ docker network create austrodb-main
$ docker network create austrodb-redis
```

## Main (postgres)


Postgres Database (`postgres:16.1-alpine`)

```sh
$ docker run -d --name austrodb-main -p 5432:5432 \
> -e POSTGRES_DB=austrody \
> -e POSTGRES_USER=penguin -e POSTGRES_PASSWORD=postgres \
> --network austrodb-main postgres:16.1-alpine
```

Dashbord (`dpage/pgadmin4`)

```sh
$ docker run -d --name pg_dashbord -p 5433:80 \
> -e PGADMIN_DEFAULT_EMAIL=penguin@austrody.sns \
> -e PGADMIN_DEFAULT_PASSWORD=penguin \
> --network austrodb-main dpage/pgadmin4
```

## Auth and Cache (Redis)

Auth Database (`redis:7.2.4-alpine`)

```sh
# pwd: gustrody/db
$ docker build -f redis.dockerfile -t austrodb-auth:1.0 \
> --secret id=redis,src=./redis_secret \
> --build-arg DB=auth .

$ docker run -d --name austordb-auth -p 6739:6739 \
> --network austrodb-redis austrodb-auth:1.0
```