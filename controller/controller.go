package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"./control"
)

// SET GO111MODULE=off

func main() {

	dummyMode := false
	configName := ""
	cmdMode := control.RunCmdMode

	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	runFlgDummy := runCmd.Bool("dummy", false, "Use dummy configuration")
	runFlgConfig := runCmd.String("name", "configuration.xml", "XML configuration file to load")

	configCmd := flag.NewFlagSet("config", flag.ExitOnError)
	configFlgDummy := configCmd.Bool("dummy", false, "Use dummy configuration")
	configFlgConfig := configCmd.String("name", "configuration.xml", "XML configuration name to save configuration")

	if len(os.Args) < 2 {
		fmt.Println("expected 'run' or 'config' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		runCmd.Parse(os.Args[2:])
		dummyMode = *runFlgDummy
		configName = *runFlgConfig
	case "config":
		configCmd.Parse(os.Args[2:])
		dummyMode = *configFlgDummy
		configName = *configFlgConfig
		cmdMode = control.ConfigCmdMode
		//os.Exit(1)
	}

	// flag.Parse()
	if *runFlgDummy == true {
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
	controller.InitController(&regDevices, &logger, cmdMode, configName, dummyMode)

	if cmdMode != control.ConfigCmdMode {
		controller.OnStart()
		controller.Run()
	}

}
