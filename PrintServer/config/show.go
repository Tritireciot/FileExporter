package config

import (
	"fmt"
	"log"
)

func Show() {
	log.Print(GetString())
}

func GetString() (str string) {
	if len(MQAccess) > 0 {
		str += "Message queue paths:\n"
		for i := range MQAccess {
			str += fmt.Sprintf("\tPath: %s\n\tPort: %d\n", MQAccess[i].Path, MQAccess[i].Port)
		}
	}

	for _, cfg := range DbConfig {
		str += cfg.String()
	}

	if len(SubAccess) > 0 {
		str += fmt.Sprintf("Sub Access:")
		for _, cfg := range SubAccess {
			str += "\n\t" + cfg.GetPath()
		}
	}

	return str
}
