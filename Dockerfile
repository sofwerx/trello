FROM debian:sid

RUN apt-get update
RUN apt-get install -y --no-install-recommends \
		ca-certificates \
		curl \
		netbase \
		wget

RUN set -ex; \
	if ! command -v gpg > /dev/null; then \
		apt-get install -y --no-install-recommends \
			gnupg \
			dirmngr \
		; \
	fi

# procps is very common in build systems, and is a reasonably small package
RUN apt-get install -y --no-install-recommends \
		bzr \
		git \
		mercurial \
		openssh-client \
		subversion \
		\
		procps

# gcc for cgo
RUN apt-get install -y --no-install-recommends \
		g++ \
		gcc \
		libc6-dev \
		make \
		pkg-config

ENV GOLANG_VERSION 1.10.3

RUN set -eux; \
	\
# this "case" statement is generated via "update.sh"
	dpkgArch="$(dpkg --print-architecture)"; \
	case "${dpkgArch##*-}" in \
		amd64) goRelArch='linux-amd64'; goRelSha256='fa1b0e45d3b647c252f51f5e1204aba049cde4af177ef9f2181f43004f901035' ;; \
		armhf) goRelArch='linux-armv6l'; goRelSha256='d3df3fa3d153e81041af24f31a82f86a21cb7b92c1b5552fb621bad0320f06b6' ;; \
		arm64) goRelArch='linux-arm64'; goRelSha256='355128a05b456c9e68792143801ad18e0431510a53857f640f7b30ba92624ed2' ;; \
		i386) goRelArch='linux-386'; goRelSha256='3d5fe1932c904a01acb13dae07a5835bffafef38bef9e5a05450c52948ebdeb4' ;; \
		ppc64el) goRelArch='linux-ppc64le'; goRelSha256='f3640b2f0990a9617c937775f669ee18f10a82e424e5f87a8ce794a6407b8347' ;; \
		s390x) goRelArch='linux-s390x'; goRelSha256='34385f64651f82fbc11dc43bdc410c2abda237bdef87f3a430d35a508ec3ce0d' ;; \
		*) goRelArch='src'; goRelSha256='567b1cc66c9704d1c019c50bef946272e911ec6baf244310f87f4e678be155f2'; \
			echo >&2; echo >&2 "warning: current architecture ($dpkgArch) does not have a corresponding Go binary release; will be building from source"; echo >&2 ;; \
	esac; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.${goRelArch}.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "${goRelSha256} *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	\
	if [ "$goRelArch" = 'src' ]; then \
		echo >&2; \
		echo >&2 'error: UNIMPLEMENTED'; \
		echo >&2 'TODO install golang-any from jessie-backports for GOROOT_BOOTSTRAP (and uninstall after build)'; \
		echo >&2; \
		exit 1; \
	fi; \
	\
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH

# Install facetrack/tellobot/opencv/sdl/ffmpeg dependencies

RUN apt-get install -y \
    libopencv-dev python-opencv git cmake build-essential \
    libsdl2-dev libsdl2-image-dev libsdl2-ttf-dev libsdl2-mixer-dev \
    ffmpeg libavcodec-dev libavdevice-dev libavfilter-dev libavformat-dev libavresample-dev \
    libavutil-dev libpostproc-dev libswscale-dev

RUN go get -d -u gobot.io/x/gobot/...
RUN go get -u -d gocv.io/x/gocv
RUN bash -c 'go get -v github.com/veandco/go-sdl2/{sdl,img,mix,ttf}'

RUN cd $GOPATH/src/gobot.io/x/gobot \
 && git checkout dev \
 && git fetch

ADD facetrackingnojs.go .
#ADD facetracking.go .
#ADD tello_control.go .

#CMD go run tello_control.go
CMD go run facetrackingnojs.go proto.txt model

