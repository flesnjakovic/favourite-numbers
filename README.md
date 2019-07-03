# Favourite number

This repository contains code for 'Favourite numbers' application. Every user (identified by a username) has a favourite number. The system maintains a list of users and their favourite numbers, and exposes a single websocket endpoint. The websocket accepts two types of messages:

1. a message to set a user's favourite number
2. a message to list all users (sorted alphabetically) and their favourite numbers

The websocket has one type of response message:

1. the alphabetical listing of all known users and their favourite number.
