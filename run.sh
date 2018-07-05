#!/bin/bash -x
#exec go run tello_control.go
exec go run facetrackingnojs.go proto.txt model ${LANDTIME:-300}
