package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"./control"
)

func main() {

	flgDummy := flag.Bool("dummy", false, "Use dummy configuration")
	flgConfig := flag.String("config", "configuration.xml", "XML configuration file to load")
	flag.Parse()
	if *flgDummy == true {
		fmt.Println("Dummy Configutation used")
	}

	regDevices := control.RegDevices{

		"TempSensor":      reflect.TypeOf(control.TempSensor{}),
		"DummyTempSensor": reflect.TypeOf(control.DummyTempSensor{}),
		"DummyRelay":      reflect.TypeOf(control.DummyRelay{}),
		"SimpleRelay":     reflect.TypeOf(control.SimpleRelay{}),
		"SimpleSSR":       reflect.TypeOf(control.SimpleSSR{}),
		"SimpleRIMM":      reflect.TypeOf(control.SimpleRIMM{}),
		"ActiveBuzzer":    reflect.TypeOf(control.ActiveBuzzer{}),
		"DummyBuzzer":     reflect.TypeOf(control.DummyBuzzer{}),
	}

	fmt.Println("Starting Controller...")

	logger := control.Logger{}
	logger.Init()
	logger.SetDebug(true)
	logger.Add("default", control.LogLevelAll, os.Stdout)

	controller := control.Control{}
	controller.InitController(&regDevices, &logger, *flgConfig, *flgDummy)
	controller.Run()

}
