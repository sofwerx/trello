# docker-tello-facetrack

This is javaguy's tello facetrack, as forked by seahope, as documented in this tellopilots.com thread:

- https://tellopilots.com/threads/face-tracking-with-tello-and-gocv.374/

This repository is an attempt at a dockerized version of this project.

## Usage:

This requires an X11 display to run.

Please read the [GLX.md](GLX.md) file for instructions for your particular X11 display.

Note: You may need to run `xterm` or some other X11 client app once on your mac to generate the `MIT_MAGIC_COOKIE` in `~/.Xauthority` for your XQuartz display.

Ensure your docker-engine hosting machine is associated via Wifi as a client to the tello drone, then run:

    make

If you change anything, you will need to force a new build:

    make build run

