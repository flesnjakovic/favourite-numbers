# Favourite number

This repository contains code for 'Favourite numbers' application. Every user (identified by a username) has a favourite number. The system maintains a list of users and their favourite numbers, and exposes a single websocket endpoint. The websocket accepts two types of messages:

1. a message to set a user's favourite number
2. a message to list all users (sorted alphabetically) and their favourite numbers

The websocket has one type of response message:

1. the alphabetical listing of all known users and their favourite number.

There are 2 types of workers that can be used interchangeably. Go lang and python implementation. By default, `docker-compose.yml` uses go worker. If needed, python worker can be referenced in `docker-compose.yml` like:

```yaml
worker:
  image: flesnjakovic/favourite-number-python-worker:stable
```

`docker-compose-dev.yml` file can be used for development and it builds docker images from worker and server folders.