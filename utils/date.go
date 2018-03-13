package utils

import (
	"fmt"
	"strings"
	"time"
)

type dateMonInterface interface {
	Sync(ch chan string)
}

type dateMon struct {
}

var dconfig struct {
	Date struct {
		DateChar   string
		DateFormat string
		Interval   int
	}
}
var dConfig = &dconfig.Date

// Initialise config and create a new date monitor
func NewDateMonitor(rawConf []byte) (dateMon, error) {

	err := GetConfig(rawConf, &dconfig)

	return dateMon{}, err
}

func (d *dateMon) Sync(ch chan string) {
	for t := time.Tick(time.Second * time.Duration(dConfig.Interval)); ; <-t {

		msg := []string{dConfig.DateChar}

		c := time.Now()
		y, m, d := c.Date()
		hour, min, _ := c.Clock()
		msg = append(msg, fmt.Sprintf("%d-%s-%d %2d:%02d", d, m, y, hour, min))

		ch <- strings.Join(msg, " ")
	}
}
