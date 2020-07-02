package common

import (
	"os"
	log "github.com/sirupsen/logrus" 
)

func CheckError(err error) {
	if err != nil {
		log.Errorf("Fatal error：%s", err.Error())
		os.Exit(1)
	}
}
