FROM golang:stretch
RUN apt-get update

# Install facetrack/tellobot/opencv/sdl/ffmpeg dependencies
RUN apt-get install -y \
    libopencv-dev python-opencv git cmake build-essential \
    libsdl2-dev libsdl2-image-dev libsdl2-ttf-dev libsdl2-mixer-dev \
    ffmpeg libavcodec-dev libavdevice-dev libavfilter-dev libavformat-dev libavresample-dev \
    libswscale-dev libtbb2 libtbb-dev libjpeg-dev libpng-dev libtiff-dev libdc1394-22-dev \
    libavutil-dev libpostproc-dev unzip libgtk2.0-dev pkg-config

RUN apt-get install -y sudo

RUN go get -u -d gocv.io/x/gocv
RUN cd $GOPATH/src/gocv.io/x/gocv \
 && make deps \
 && make install

RUN bash -c 'go get -v github.com/veandco/go-sdl2/{sdl,img,mix,ttf}'

RUN go get -d -u gobot.io/x/gobot/...
RUN go get github.com/gobuffalo/uuid
RUN cd $GOPATH/src/gobot.io/x/gobot \
 && git checkout dev \
 && git fetch

ENV LANDTIME=300

ADD proto.txt .
ADD model .
ADD facetrackingnojs.go .
#ADD facetracking.go .
#ADD tello_control.go .

ADD run.sh /run.sh

CMD /run.sh
