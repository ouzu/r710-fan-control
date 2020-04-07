package main

import (
	"flag"
	"fmt"
	"github.com/md14454/gosensors"
	"log"
	"os/exec"
	"regexp"
	"time"
)

var ipmiArgs = []string{"-I", "lanplus", "-H", "[address here...]", "-U", "root", "-P", "[password here...]"}

var ipmiReset = []string{"raw", "0x30", "0x30", "0x01", "0x01"}
var ipmiManual = []string{"raw", "0x30", "0x30", "0x01", "0x00"}

var ipmiSpeed = []string{"raw", "0x30", "0x30", "0x02", "0xff"}

type sensorList []gosensors.SubFeature

func getSensors() sensorList {
	sensors := make(sensorList, 0)

	tempRe := regexp.MustCompile(`.*_input`)
	coreRe := regexp.MustCompile(`Core.*`)

	for _, chip := range gosensors.GetDetectedChips() {
		for _, feature := range chip.GetFeatures() {
			if coreRe.MatchString(feature.GetLabel()) {
				for _, subfeature := range feature.GetSubFeatures() {
					if tempRe.MatchString(subfeature.Name) {
						sensors = append(sensors, subfeature)
					}
				}
			}
		}
	}

	return sensors
}

func getTemps() []float64 {
	sensors := getSensors()

	tempList := make([]float64, len(sensors))

	for i, sensor := range sensors {
		tempList[i] = sensor.GetValue()
	}

	return tempList
}

func maxTemp() float64 {
	gosensors.Init()
	defer gosensors.Cleanup()

	max := 0.0
	for _, t := range getTemps() {
		if t > max {
			max = t
		}
	}

	return max
}

func ipmiCall(command []string) {
	cmd := exec.Command("ipmitool", append(ipmiArgs, command...)...)
	_, err := cmd.Output()
	if err != nil {
		log.Fatalln("error calling ipmitool:", err)
	}
}

func autoFanMode() {
	ipmiCall(ipmiReset)
}

func manualFanMode() {
	ipmiCall(ipmiManual)
}

func setFan(percent int) {
	ipmiCall(append(ipmiSpeed, fmt.Sprintf("%#02x", percent)))
}

func main() {
	speed := flag.Int("speed", -1, "set manual speed (in %)")
	curve := flag.Int("curve", 5, "curve to use")
	mode := flag.String("mode", "print", "mode (print, reset, manual or auto)")
	debug := flag.Bool("debug", false, "print debug output")

	flag.Parse()

	if *debug {
		log.Println("mode:", *mode)
		log.Println("speed:", *speed)
	}

	switch *mode {
	case "print":
		fmt.Println("highest core temperature:", maxTemp())
	case "reset":
		autoFanMode()
	case "manual":
		if *speed < 0 {
			fmt.Println("please specify a speed")
			return
		}

		if *speed > 100 {
			fmt.Println("speed must be smaller than 100")
			return
		}

		manualFanMode()
		setFan(*speed)
	case "auto":
		manualFanMode()
		defer autoFanMode()

		curr := maxTemp()

		history := make([]float64, 4)

		for i := range history {
			history[i] = curr
		}

		avg := curr

		speed := 20
		lastSpeed := 19

		for {
			history[0] = history[1]
			history[1] = history[2]
			history[2] = history[3]
			history[3] = curr

			curr = maxTemp()

			avg = (history[0] + history[1] + history[2] + history[3] + curr) / 5

			if *debug {
				fmt.Println("current temperature:", curr)
				fmt.Println("floating average:", avg)
			}

			lastSpeed = speed
			speed = 20

			if curr > avg*1.1 {
				fmt.Println("Surge detected")
				avg = curr
			}

			if curr > 70 {
				fmt.Println("Temperature >70 !!!")
				autoFanMode()
				time.Sleep(time.Minute)
				manualFanMode()
				continue
			} else if avg < 25 {
				speed = 0
			} else {
				switch *curve {
				case 1:
					// 0 at 25, 100 at 80
					speed = int(20 * (avg - 25) / 11)
				case 2:
					// 0 at 25, 15 at 40, 100 at 80
					speed = int((((9 * (avg - 40)) / 440) + 1) * (avg - 25))
				case 3:
					// 0 at 25, 10 at 40, 100 at 80
					speed = int(((19*(avg-40))/660 + 1) * (avg - 25))
				case 4:
					// 0 at 25, 15 at 45, 100 at 80
					speed = int(((47*(avg-45))/1540 + 0.75) * (avg - 25))
				case 5:
					// 0 at 25, 10 at 45, 100 at 80
					speed = int(((29*(avg-45))/770 + 0.5) * (avg - 25))
				}
			}

			if speed != lastSpeed {
				if *debug {
					fmt.Println("setting speed to", speed)
				}

				if speed >= 0 && speed <= 100 {
					setFan(speed)
				} else {
					fmt.Println("invalid value")
				}

			}

			time.Sleep(2 * time.Second)
		}
	}
}
