package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"./control"
)

// SET GO111MODULE=off

const (
	DefaultSensorCount = 3
	DefaultRelayCount  = 3
	DefaultSSRCount    = 1
	DefaultBuzzerCount = 1
)

func main() {

	dummyMode := false
	debugMode := false
	configName := ""
	cmdMode := control.RunCmdMode
	CmdInfo := control.CmdInfo{}

	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	runFlgDummy := runCmd.Bool("dummy", false, "Use dummy configuration")
	runFlgDebug := runCmd.Bool("debug", false, "Run in debug mode")
	runFlgConfig := runCmd.String("name", "configuration.xml", "XML configuration file to load")

	configCmd := flag.NewFlagSet("config", flag.ExitOnError)
	configFlgDummy := configCmd.Bool("dummy", false, "Use dummy configuration")
	configFlgDebug := configCmd.Bool("debug", false, "Run in debug mode")
	configFlgName := configCmd.String("name", "configuration.xml", "XML configuration name to save configuration")
	configFlgList := configCmd.Bool("list", false, "List 64 bit addreeses for 1-wire devices available then exit")
	configFlgSens1 := configCmd.String("sens1", "unknown", "Set sensor 1. format \"<name[:<net address>\". Default \"Temp Sensor 1\"")
	configFlgSens2 := configCmd.String("sens2", "unknown", "Set sensor 2. format \"<name[:<net address>\". Default \"Temp Sensor 2\"")
	configFlgSens3 := configCmd.String("sens3", "unknown", "Set sensor 3. format \"<name[:<net address>\". Default \"Temp Sensor 3\"")

	if len(os.Args) < 2 {
		fmt.Println("expected 'run' or 'config' subcommands")
		os.Exit(1)
	}

	sensors := []string{"Temp Sensor 1", "Temp Sensor 2", "Temp Sensor 3"}
	mode := os.Args[1]

	switch mode {
	case "run":
		runCmd.Parse(os.Args[2:])
		dummyMode = *runFlgDummy
		configName = *runFlgConfig
		debugMode = *runFlgDebug
	case "config":
		configCmd.Parse(os.Args[2:])
		dummyMode = *configFlgDummy
		configName = *configFlgName
		debugMode = *configFlgDebug
		cmdMode = control.ConfigCmdMode
		tempSensors := []string{*configFlgSens1, *configFlgSens2, *configFlgSens3}
		f := true
		for i := 0; i < DefaultSensorCount; i++ {
			if tempSensors[i] != "default" {
				f = false
				break
			}
		}

		for i := 0; i < DefaultSensorCount; i++ {
			if !f {
				if tempSensors[i] != "unknown" {
					CmdInfo.Sensors = append(CmdInfo.Sensors, tempSensors[i])
					continue
				}
				CmdInfo.Sensors = append(CmdInfo.Sensors, "")
			} else {
				CmdInfo.Sensors = append(CmdInfo.Sensors, fmt.Sprintf("Temp Sensor %d", i+1))
			}
		}
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
	logger.SetDebug(debugMode)
	logger.Add("default", control.LogLevelAll, os.Stdout)

	if mode == "config" {
		sensors[0] = *configFlgSens1
		sensors[1] = *configFlgSens2
		sensors[2] = *configFlgSens3
	}

	if *configFlgList == true {
		netAddresses, errlist := control.GetActiveNetlinkAddresses(&logger)
		if errlist != nil {
			fmt.Printf("error reading net addresses: %s", errlist)
		} else {
			for i, addrs := range netAddresses {
				fmt.Printf("(%d) = %d", i, addrs)
				logger.LogMessage("(%d) = %d", i, addrs)
			}
		}
		os.Exit(0)
	}

	controller := control.Control{}
	controller.InitController(&regDevices, &logger, sensors, cmdMode, configName, dummyMode)

	if cmdMode != control.ConfigCmdMode {
		controller.OnStart()
		controller.Run()
	}

}
