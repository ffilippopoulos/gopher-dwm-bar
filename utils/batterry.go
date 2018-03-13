package utils

import (
	"fmt"
	"github.com/distatus/battery"
	"log"
	"strings"
	"time"
)

type batteryMonInterface interface {
	update() error
	Sync(ch chan string)
}

type batteryMon struct {
	state         string
	capacity      int
	timeRemaining float64
}

var bconfig struct {
	Battery struct {
		CapacityCharList []string
		ChargingChar     string
		FullChar         string
		Interval         int
	}
}
var bConfig = &bconfig.Battery

// Initialise config and create a new batery monitor
func NewBatteryMonitor(rawConf []byte) (batteryMon, error) {

	err := GetConfig(rawConf, &bconfig)

	return batteryMon{}, err
}

func (b *batteryMon) update() error {

	//Lets assume we only have the main battery
	bat, err := battery.Get(0)
	if err != nil {
		return err
	}

	// states: ["Unknown", "Empty", "Full", "Charging", "Discharging"]
	b.state = bat.State.String()

	capacity := bat.Current / bat.Full
	b.capacity = int(capacity * 100)

	b.timeRemaining = bat.Current / bat.ChargeRate // in hours

	return nil
}

// Update every interval and write to channel
func (b *batteryMon) Sync(ch chan string) {
	for t := time.Tick(time.Second * time.Duration(bConfig.Interval)); ; <-t {

		if err := b.update(); err != nil {
			log.Println("Updating battery details failed: %v", err)
		}

		var msg []string

		if b.state == "Discharging" {

			step := 100 / len(bConfig.CapacityCharList)
			c := bConfig.CapacityCharList[b.capacity/step-1]
			//msg = append(msg, fmt.Sprintf("$(echo -e \"\\u%s\")", c))
			msg = append(msg, c)

		} else if b.state == "Charging" {

			//msg = append(msg, fmt.Sprintf("$(echo -e \"\\u%s\")", bConfig.ChargingChar))
			msg = append(msg, bConfig.ChargingChar)

		} else if b.state == "Full" {

			//msg = append(msg, fmt.Sprintf("$(echo -e \"\\u%s\")", bConfig.FullChar))
			msg = append(msg, bConfig.FullChar)

		} else {

			log.Println("Else")

		}

		msg = append(msg, fmt.Sprintf("%d%%", b.capacity))
		if b.state == "Discharging" {
			msg = append(msg, fmt.Sprintf("(%.2fh)", b.timeRemaining))
		}

		ch <- strings.Join(msg, " ")
	}
}
