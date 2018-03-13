package utils

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type AudioMixerDevice string

const (
	Master    AudioMixerDevice = "Master"
	Headphone AudioMixerDevice = "Headphone"
)

type AudioDeviceStatus string

const (
	On  AudioDeviceStatus = "on"
	Off AudioDeviceStatus = "off"
)

var execCommand = exec.Command

type audioMonInterface interface {
	update() error
	Sync(ch chan string)
}

type audioMon struct {
	activeDevice AudioMixerDevice
	status       AudioDeviceStatus
	volume       string
}

var aconfig struct {
	Audio struct {
		VolumeMuteChar   string
		VolumeZeroChar   string
		VoluemeOnChar    string
		HeadphonesOnChar string
		Interval         int
	}
}
var aConfig = &aconfig.Audio

// Initialise config and create a new batery monitor
func NewAudioMonitor(rawConf []byte) (audioMon, error) {

	err := GetConfig(rawConf, &aconfig)

	return audioMon{}, err
}

// based on amixer, et.c. line: `Mono: Playback 53 [72%] [-21.00dB] [on]`
//var re = regexp.MustCompile(`.*Playback\s+\d+\d+\s+\[(?P<volume>.*)\]\s+\[.*\]\s+\[(?P<status>\w+)\]`)
var re = regexp.MustCompile(`.*Playback\s+.*\s+\[(?P<volume>.*)\]\s+\[.*\]\s+\[(?P<status>\w+)\]`)
var vTemplate = []byte("$volume")
var sTemplate = []byte("$status")

// Returns status [on/off] and volume for the Master
func getMixerData(device AudioMixerDevice) (AudioDeviceStatus, string, error) {

	args := []string{"amixer", "get", fmt.Sprintf("%s", device)}

	out, err := execCommand(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		return Off, "", err
	}

	// To store the results
	result := make(map[string]string)

	// Iterate through matches and store results
	for _, submatches := range re.FindAllSubmatchIndex(out, -1) {
		volume := []byte{}
		volume = re.Expand(volume, vTemplate, out, submatches)
		status := []byte{}
		status = re.Expand(status, sTemplate, out, submatches)
		result[string(status)] = string(volume)
	}

	if len(result) > 1 {
		// Means that we have channels that are on and others that are off
		return Off, "", errors.New("missmatch on mono channels")
	} else if len(result) == 0 {
		log.Println("no channels found")
		return Off, "", errors.New("no channels found")
	} else {
		for s, v := range result {
			if s == "on" {
				return On, v, nil
			} else {
				return Off, v, nil
			}
		}
	}

	return Off, "", errors.New("unreachable")
}

func (a *audioMon) update() error {
	mStatus, mVolume, err := getMixerData(Master)
	if err != nil {
		return err
	}
	// Update general status and volume)
	a.status = mStatus
	a.volume = mVolume

	// cCheck weather headphone is enabled
	hStatus, _, _ := getMixerData(Headphone)
	//if err != nil {
	//	return err
	//}
	if hStatus == On {
		a.activeDevice = Headphone
	} else {
		a.activeDevice = Master
	}

	return nil
}

var audioMsg string

// Update every interval and write to channel
func (a *audioMon) Sync(ch chan string) {
	for t := time.Tick(time.Second * time.Duration(aConfig.Interval)); ; <-t {
		err := a.update()
		if err != nil {
			log.Println("%v", err)
		}

		msg := []string{}
		if a.status == Off {
			msg = append(msg, aConfig.VolumeMuteChar)
		} else if a.status == On {
			if a.activeDevice == Headphone {
				msg = append(msg, aConfig.HeadphonesOnChar)
			} else if a.activeDevice == Master {
				if a.volume != "0%" {
					msg = append(msg, aConfig.VoluemeOnChar)
				} else {
					msg = append(msg, aConfig.VolumeZeroChar)
				}
			} else {
				log.Println("Unknown Device")
			}
		} else {
			log.Println("Unknown Status")
		}

		msg = append(msg, a.volume)

		newAudioMsg := strings.Join(msg, " ")
		if newAudioMsg != audioMsg {
			audioMsg = newAudioMsg
			ch <- audioMsg
		}
	}

}
