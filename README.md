# docker-tello-facetrack

This is javaguy's tello facetrack, as forked by seahope, as documented in this tellopilots.com thread:

- https://tellopilots.com/threads/face-tracking-with-tello-and-gocv.374/

This repository is an attempt at a dockerized version of this project.

## Usage:

Ensure your docker-engine hosting machine is associated via Wifi as a client to the tello drone, then run:

    docker-compose up

If you change anything but the changes do not appear to happen inside the docker container, use this:

    docker-compose build
    docker-compose up --force-recreate

