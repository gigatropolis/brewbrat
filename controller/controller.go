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

		"TempSensor":      reflect.TypeOf(control.TempSensor{}),
		"DummyTempSensor": reflect.TypeOf(control.DummyTempSensor{}),
	}

	fmt.Println("Starting Controller...")

	logger := control.Logger{}
	logger.Init()
	logger.SetDebug(true)
	logger.Add("default", control.LogLevelAll, os.Stdout)

	var t1 control.ISensor = &control.TempSensor{}

	props := []control.Property{
		{Name: "Name", PropType: "string", Hidden: false, Value: "temp Sensor 1", Comment: "Sensor Name", Choice: ""},
		{Name: "Address", PropType: "string", Hidden: false, Value: uint64(7205759448148251176), Comment: "1-Wire sensor address", Choice: ""},
		{Name: "Units", PropType: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""},
	}
	chnSensorValue := make(chan control.SensorMessage)

	t1.InitSensor(&logger, props, chnSensorValue)
	t1.OnStart()
	go t1.Run()

	for sensor := range chnSensorValue {
		fmt.Printf("Recieved Value %.3f F", sensor.Value)
		time.Sleep(time.Second * 3)
	}

}
