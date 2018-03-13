package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"
)

func TestAudio(t *testing.T) {
	conf, err := ioutil.ReadFile("../configuration.json")
	if err != nil {
		log.Fatal(err)
	}
	a, _ := NewAudioMonitor(conf)
	err = a.update()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(a.status, a.volume, a.activeDevice)
}
