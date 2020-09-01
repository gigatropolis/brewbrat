package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"./control"
)

var RegSensors map[string]reflect.Type

func main() {

	RegSensors = map[string]reflect.Type{

		"TempSensor":      reflect.TypeOf(TempSensor{}),
		"DummyTempSensor": reflect.TypeOf(DummyTempSensor{}),
	}

	fmt.Println("Starting Controller...")

	logger := control.Logger{}
	logger.Init()
	logger.SetDebug(true)
	logger.Add("default", control.LogLevelAll, os.Stdout)

	var t1 control.ISensor = &control.TempSensor{}

	props := []control.Property{
		{"Name", "string", "temp Sensor 1", "Sensor Name"},
		{"Address", "string", uint64(7205759448148251176), "1-Wire sensor address"},
		{"Units", "string", "°F", "Units for Sensor"},
	}

	t1.Init(&logger, props)
	t1.OnStart()

	active := true
	for active {
		value, err := t1.OnRead()
		if err != nil {
			fmt.Println("can't read sensor")
			active = false
		} else {
			t1.LogMessage("Sensor value = %.3f%s", value, t1.GetUnits())
		}
		time.Sleep(time.Second * 3)
	}
}
