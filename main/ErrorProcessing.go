package main

import (
	log "github.com/sirupsen/logrus"
)

//Обработка ошибок
func CheckError(err error) {
	if err != nil {
		//panic(err)
		log.Error(err)
	}
}
