package main

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"time"

	"gobot.io/x/gobot/platforms/joystick"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gocv.io/x/gocv"
)

const maxJoyVal = 32768
const frameSize = frameX * frameY * 3
const frameX = 400
const frameY = 300

var drone = tello.NewDriver("8890")
var window = gocv.NewWindow("Tello")

var ffmpeg = exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0",
	"-pix_fmt", "bgr24", "-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1")
var ffmpegIn, _ = ffmpeg.StdinPipe()
var ffmpegOut, _ = ffmpeg.StdoutPipe()

var joyAdaptor = joystick.NewAdaptor()
var stick = joystick.NewDriver(joyAdaptor, "dualshock4")
var flightData *tello.FlightData
var tracking = false
var detectSize = false
var distTolerance = 0.05 * dist(0, 0, frameX, frameY)

func init() {
	handleJoystick()
	go func() {
		if err := ffmpeg.Start(); err != nil {
			fmt.Println(err)
			return
		}

		drone.On(tello.FlightDataEvent, func(data interface{}) {
			flightData = data.(*tello.FlightData)
		})

		drone.On(tello.ConnectedEvent, func(data interface{}) {
			fmt.Println("Connected")
			drone.StartVideo()
			drone.SetVideoEncoderRate(tello.VideoBitRateAuto)
			drone.SetExposure(0)
			gobot.Every(100*time.Millisecond, func() {
				drone.StartVideo()
			})
		})

		drone.On(tello.VideoFrameEvent, func(data interface{}) {
			pkt := data.([]byte)
			if _, err := ffmpegIn.Write(pkt); err != nil {
				fmt.Println(err)
			}
		})

		robot := gobot.NewRobot("tello",
			[]gobot.Connection{},
			[]gobot.Connection{joyAdaptor},
			[]gobot.Device{stick},
			[]gobot.Device{drone},
		)

		robot.Start()
	}()
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("How to run:\ngo run facetracking.go [protofile] [modelfile]")
		return
	}

	proto := os.Args[1]
	model := os.Args[2]

	net := gocv.ReadNetFromCaffe(proto, model)
	if net.Empty() {
		fmt.Printf("Error reading network model from : %v %v\n", proto, model)
		return
	}
	defer net.Close()

	green := color.RGBA{0, 255, 0, 0}

	if net.Empty() {
		fmt.Printf("Error reading network model from : %v %v\n", proto, model)
		return
	}
	defer net.Close()

	refDistance := float64(0)
	detected := false
	left := float32(0)
	top := float32(0)
	right := float32(0)
	bottom := float32(0)

	for {
		buf := make([]byte, frameSize)
		if _, err := io.ReadFull(ffmpegOut, buf); err != nil {
			fmt.Println(err)
			continue
		}
		img := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)
		if img.Empty() {
			continue
		}
		W := float32(img.Cols())
		H := float32(img.Rows())
		blob := gocv.BlobFromImage(img, 1.0, image.Pt(128, 96), gocv.NewScalar(104.0, 177.0, 123.0, 0), false, false)
		defer blob.Close()

		net.SetInput(blob, "data")

		detBlob := net.Forward("detection_out")
		defer detBlob.Close()

		detections := gocv.GetBlobChannel(detBlob, 0, 0)
		defer detections.Close()

		for r := 0; r < detections.Rows(); r++ {
			confidence := detections.GetFloatAt(r, 2)
			if confidence < 0.5 {
				continue
			}

			left = detections.GetFloatAt(r, 3) * W
			top = detections.GetFloatAt(r, 4) * H
			right = detections.GetFloatAt(r, 5) * W
			bottom = detections.GetFloatAt(r, 6) * H

			left = min(max(0, left), W-1)
			right = min(max(0, right), W-1)
			bottom = min(max(0, bottom), H-1)
			top = min(max(0, top), H-1)

			rect := image.Rect(int(left), int(top), int(right), int(bottom))
			gocv.Rectangle(&img, rect, green, 3)
			detected = true
		}

		window.IMShow(img)
		if window.WaitKey(10) >= 0 {
			break
		}

		if !tracking || !detected {
			continue
		}

		if detectSize {
			detectSize = false
			refDistance = dist(left, top, right, bottom)
		}

		distance := dist(left, top, right, bottom)

		if right < W/2 {
			drone.CounterClockwise(50)
		} else if left > W/2 {
			drone.Clockwise(50)
		} else {
			drone.Clockwise(0)
		}

		if top < H/10 {
			drone.Up(25)
		} else if bottom > H-H/10 {
			drone.Down(25)
		} else {
			drone.Up(0)
		}

		if distance < refDistance-distTolerance {
			drone.Forward(20)
		} else if distance > refDistance+distTolerance {
			drone.Backward(20)
		} else {
			drone.Forward(0)
		}
	}
}

func dist(x1, y1, x2, y2 float32) float64 {
	return math.Sqrt(float64((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1)))
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func handleJoystick() {
	stick.On(joystick.CirclePress, func(data interface{}) {
		drone.Forward(0)
		drone.Up(0)
		drone.Clockwise(0)
		tracking = !tracking
		if tracking {
			detectSize = true
			println("tracking")
		} else {
			detectSize = false
			println("not tracking")
		}
	})
	stick.On(joystick.SquarePress, func(data interface{}) {
		fmt.Println("battery:", flightData.BatteryPercentage)
	})
	stick.On(joystick.TrianglePress, func(data interface{}) {
		drone.TakeOff()
		println("Takeoff")
	})
	stick.On(joystick.XPress, func(data interface{}) {
		drone.Land()
		println("Land")
	})
	stick.On(joystick.RightY, func(data interface{}) {
		val := float64(data.(int16))
		if val >= 0 {
			drone.Backward(tello.ValidatePitch(val, maxJoyVal))
		} else {
			drone.Forward(tello.ValidatePitch(val, maxJoyVal))
		}
	})
	stick.On(joystick.RightX, func(data interface{}) {
		val := float64(data.(int16))
		if val >= 0 {
			drone.Right(tello.ValidatePitch(val, maxJoyVal))
		} else {
			drone.Left(tello.ValidatePitch(val, maxJoyVal))
		}
	})
	stick.On(joystick.LeftY, func(data interface{}) {
		val := float64(data.(int16))
		if val >= 0 {
			drone.Down(tello.ValidatePitch(val, maxJoyVal))
		} else {
			drone.Up(tello.ValidatePitch(val, maxJoyVal))
		}
	})
	stick.On(joystick.LeftX, func(data interface{}) {
		val := float64(data.(int16))
		if val >= 0 {
			drone.Clockwise(tello.ValidatePitch(val, maxJoyVal))
		} else {
			drone.CounterClockwise(tello.ValidatePitch(val, maxJoyVal))
		}
	})
}
