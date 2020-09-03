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

	sensorsDefined := []control.SensorDefinition{
		{Name: "temp Sensor 1",
			Type: "TempSensor",
			Properties: []control.Property{
				{Name: "Name", PropType: "string", Hidden: false, Value: "temp Sensor 1", Comment: "Sensor Name", Choice: ""},
				{Name: "Address", PropType: "string", Hidden: false, Value: uint64(7205759448148251176), Comment: "1-Wire sensor address", Choice: ""},
				{Name: "Units", PropType: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""}},
		},
		{Name: "Dummy Temp 1",
			Type: "DummyTempSensor",
			Properties: []control.Property{
				{Name: "Name", PropType: "string", Hidden: false, Value: "Dummy Temp 1", Comment: "Sensor Name", Choice: ""},
				{Name: "Units", PropType: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""}},
		},
	}

	chnSensorValue := make(chan control.SensorMessage)
	sensors := make(map[string]control.ISensor)

	for _, sensor := range sensorsDefined {
		if _, ok := regSensors[sensor.Type]; ok {
			t1 := reflect.New(regSensors[sensor.Type]).Interface().(control.ISensor)
			t1.InitSensor(&logger, sensor.Properties, chnSensorValue)
			sensors[sensor.Name] = t1
		}
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
