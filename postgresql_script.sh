#!/bin/sh

mkdir -p $HOME/.docker_trash_data/postgres

docker run -d \
  --rm \
  --name postgres_localhost \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_USER=postgres \
  -v $HOME/.docker_trash_data/postgres:/var/lib/postgresql/data \
  -p 127.0.0.1:5432:5432 \
  postgres:14
