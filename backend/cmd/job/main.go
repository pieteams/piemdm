package main

import (
	"piemdm/pkg/configloader"
	"piemdm/pkg/log"
)

func main() {
	v, err := configloader.Load()
	if err != nil {
		panic(err)
	}
	logger := log.NewLog(v)
	logger.Info("start")

	app, cleanup, err := newApp(v, logger)
	if err != nil {
		panic(err)
	}
	app.Run()
	defer cleanup()
}
