# Gustrody

An sns server under ActivityPub protocol, in go.

## Start Up

**Requirements**: have *Docker* installed. Use a *reserve proxy* like *Nginx* or *Apache* to forward request to the port set in `.env`.

Put your *PostgreSQL* and *Redis* password in `db/{postgres,redis}_password` files, and configure your server by edit the `.env` file.

Run start up script:

```sh
$ ./start.sh # if needed, sudo
```