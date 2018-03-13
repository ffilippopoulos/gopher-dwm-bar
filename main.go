package main

import (
	"flag"
	"fmt"
	"github.com/ffilippopoulos/gopher-dwm-bar/utils"
	"io/ioutil"
	"log"
	"os/exec"
)

var execCommand = exec.Command

func main() {

	var configPath = flag.String("c", "/tmp/configutation.json", "Location of the configuration json file")
	flag.Parse()

	conf, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Audio
	var audio string
	aMon, err := utils.NewAudioMonitor(conf)
	if err != nil {
		log.Fatal(err)
	}
	aChan := make(chan string)
	go aMon.Sync(aChan)

	// Battery
	var battery string
	bMon, err := utils.NewBatteryMonitor(conf)
	if err != nil {
		log.Fatal(err)
	}
	bChan := make(chan string)
	go bMon.Sync(bChan)

	// Date
	var date string
	dMon, err := utils.NewDateMonitor(conf)
	if err != nil {
		log.Fatal(err)
	}
	dChan := make(chan string)
	go dMon.Sync(dChan)

	for {
		select {
		case msg := <-aChan:
			audio = msg
		case msg := <-bChan:
			battery = msg
		case msg := <-dChan:
			date = msg
		}

		name := fmt.Sprintf("%s %s %s", audio, battery, date)
		args := []string{"xsetroot", "-name", name}
		_, err := execCommand(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			log.Println(err)
		}
	}

}
