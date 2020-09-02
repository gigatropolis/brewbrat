package main

import (
	"fmt"
	"os"
	"reflect"

	"./control"
)

//RegSensors stores all sensor types that can be used.
type RegSensors map[string]reflect.Type

func main() {

	regSensors := RegSensors{

		"TempSensor":      reflect.TypeOf(control.TempSensor{}),
		"DummyTempSensor": reflect.TypeOf(control.DummyTempSensor{}),
	}

	fmt.Println("Starting Controller...")

	logger := control.Logger{}
	logger.Init()
	logger.SetDebug(true)
	logger.Add("default", control.LogLevelAll, os.Stdout)

	props := []control.Property{
		{Name: "Name", PropType: "string", Hidden: false, Value: "temp Sensor 1", Comment: "Sensor Name", Choice: ""},
		{Name: "Address", PropType: "string", Hidden: false, Value: uint64(7205759448148251176), Comment: "1-Wire sensor address", Choice: ""},
		{Name: "Units", PropType: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""},
	}
	props2 := []control.Property{
		{Name: "Name", PropType: "string", Hidden: false, Value: "Dummy Temp 1", Comment: "Sensor Name", Choice: ""},
		{Name: "Units", PropType: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""},
	}
	chnSensorValue := make(chan control.SensorMessage)

	sensors := make(map[string]control.ISensor)

	if _, ok := regSensors["temp Sensor 1"]; ok {
		t1 := reflect.New(regSensors["temp Sensor 1"]).Interface().(control.ISensor)
		t1.InitSensor(&logger, props, chnSensorValue)
		sensors["temp Sensor 1"] = t1
	}
	if _, ok := regSensors["Dummy Temp 1"]; ok {
		t2 := reflect.New(regSensors["Dummy Temp 1"]).Interface().(control.ISensor)
		t2.InitSensor(&logger, props2, chnSensorValue)
		sensors["Dummy Temp 1"] = t2
	}

	//var t1 control.ISensor = &control.TempSensor{}
	//var t2 control.ISensor = &control.DummyTempSensor{}

	//sensors["temp Sensor 1"] = t1
	//sensors["Dummy Temp 1"] = t2

	for _, sensor := range sensors {

		sensor.OnStart()
		go sensor.Run()
	}

	for senMsg := range chnSensorValue {
		name := senMsg.Name
		fmt.Printf("Recieved from '%s': Value %.3f%s", name, senMsg.Value, sensors[name].GetUnits())
	}

}
