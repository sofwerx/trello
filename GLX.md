## Indirect X11 GLX

### Notes for X Window System users

The X.org Foundation leads the X Window System project and its software forms the basis for all the X packages from the various linux distributions. Starting in Xorg 1.17 (running on CentOS 6.9 and 7.x; Ubuntu version 15.04 and up and other linux distributions) indirect GLX was disabled by default, to enable indirect GLX and to allow remote visualization of applications running under docker containers that use OpenGL on your host linux box you will have to locate the executable file Xorg (which is most likely located in /usr/bin), make a copy and create a wrapper to it in which Xorg is invoked with the +iglx flag.

On CentOS this can be achieved with:

    sudo mv /usr/bin/Xorg /usr/bin/Xorg.original
    cat <<EOF > /tmp/Xorg
    #!/usr/bin/env bash
    exec /usr/bin/Xorg.original "$@" +iglx'
    EOF
    sudo mv /tmp/Xorg /usr/bin/Xorg
    sudo chmod +x /usr/bin/Xorg
    sudo chcon --type=bin_t /usr/bin/Xorg

On Ubuntu this can be achieved with:

    sudo mv /usr/bin/Xorg /usr/bin/Xorg.original
    cat <<EOF > /tmp/Xorg
    #!/bin/sh
    #
    # Execute Xorg.wrap if it exists otherwise execute Xorg directly.
    # This allows distros to put the suid wrapper in a separate package.

    basedir=/usr/lib/xorg
    if [ -x "\$basedir"/Xorg.wrap ]; then
    exec "\$basedir"/Xorg.wrap "\$@"
    else
    exec "\$basedir"/Xorg "\$@" +iglx
    fi
    EOF
    sudo mv /tmp/Xorg /usr/bin/Xorg
    sudo chmod +x /usr/bin/Xorg

The computer will need to be rebooted.

### Mac OS X

On Mac OS X, the X windows system is called XQuartz. Mac OS X 10.5 10.6 and 10.7 installed it by default, but as of 10.8 Apple has dropped dedicated support and directs users to the open source XQuartz. 

You can install XQuartz from the OS distribution media or download the most recent version from https://www.xquartz.org/

Indirect GLX was disabled by default in Xorg 1.17, starting from XQuartz version 2.7.9, to enable indirect GLX and to allow remote visualization on the cluster you will have to issue at the terminal command prompt of your mac:

    defaults write org.macosforge.xquartz.X11 enable_iglx -bool true
    defaults read org.macosforge.xquartz.X11 nolisten_tcp -bool false

See About X11 and OS X Mountain Lion from http://support.apple.com/kb/HT5293.
For Mac OS X 10.4 (Tiger): X11 should be installed from the OS media and manually started.

