package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"./control"
)

//RegSensors stores all sensor types that can be used.
type RegDevices map[string]reflect.Type

func main() {

	regDevices := RegDevices{

		"TempSensor":      reflect.TypeOf(control.TempSensor{}),
		"DummyTempSensor": reflect.TypeOf(control.DummyTempSensor{}),
		"SimpleRelay":     reflect.TypeOf(control.SimpleRelay{}),
		"SimpleSSR":       reflect.TypeOf(control.SimpleSSR{}),
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
		if _, ok := regDevices[sensor.Type]; ok {
			t1 := reflect.New(regDevices[sensor.Type]).Interface().(control.ISensor)
			t1.InitSensor(&logger, sensor.Properties, chnSensorValue)
			sensors[sensor.Name] = t1
		}
	}

	for _, sensor := range sensors {

		sensor.OnStart()
		go sensor.Run()
	}

	relayDefined := []control.ActorDefinition{
		{Name: "Relay 1",
			Type: "SimpleRelay",
			Properties: []control.Property{
				{Name: "Name", PropType: "string", Hidden: false, Value: "Relay 1", Comment: "relay Name", Choice: ""},
				{Name: "GPIO", PropType: "string", Hidden: false, Value: "P1_38", Comment: "GPIO by name", Choice: ""},
			},
		},
		{Name: "Relay 2",
			Type: "SimpleRelay",
			Properties: []control.Property{
				{Name: "Name", PropType: "string", Hidden: false, Value: "Relay 2", Comment: "relay Name", Choice: ""},
				{Name: "GPIO", PropType: "string", Hidden: false, Value: "P1_40", Comment: "GPIO by name", Choice: ""},
			},
		},
		{Name: "SSR 1",
			Type: "SimpleSSR",
			Properties: []control.Property{
				{Name: "Name", PropType: "string", Hidden: false, Value: "SSR 1", Comment: "SSR Name", Choice: ""},
				{Name: "GPIO", PropType: "string", Hidden: false, Value: "P1_36", Comment: "GPIO by name", Choice: ""},
			},
		},
	}

	actors := make(map[string]control.IActor)

	for _, actor := range relayDefined {
		if _, ok := regDevices[actor.Type]; ok {
			t1 := reflect.New(regDevices[actor.Type]).Interface().(control.IActor)
			t1.Init(&logger, actor.Properties)
			actors[actor.Name] = t1
		}
	}

	//	for senMsg := range chnSensorValue {
	//		name := senMsg.Name
	//		fmt.Printf("Recieved from '%s': Value %.3f%s\n", name, senMsg.Value, sensors[name].GetUnits())
	//	}

	t := time.NewTicker(5000 * time.Millisecond)
	state := true

	for true {
		select {
		case senMsg := <-chnSensorValue:
			name := senMsg.Name
			fmt.Printf("Recieved from '%s': Value %.3f%s\n", name, senMsg.Value, sensors[name].GetUnits())
		case <-t.C:
			for _, act := range actors {
				if state {
					act.On()
				} else {
					act.Off()
				}
				time.Sleep(time.Millisecond * 250)
			}
			if state {
				state = false
			} else {
				state = true
			}
		}
	}

}
