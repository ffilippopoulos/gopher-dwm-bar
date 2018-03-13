package utils

import (
	"encoding/json"
)

func GetConfig(rawConf []byte, configuration interface{}) error {

	err := json.Unmarshal(rawConf, &configuration)
	if err != nil {
		return err
	}
	return nil
}
