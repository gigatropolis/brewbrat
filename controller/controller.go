package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"./www/cmd/server"

	"./config"
	"./control"
)

func HandleWebMessage(msg server.ServerCommand, sensValues SensorValues) {

	name := strings.ReplaceAll(msg.DeviceName, "_", " ")
	switch msg.Cmd {
	case server.CmdSetRelay:
		relay, ok := actors[name]
		if ok {
			sVal := string(msg.Value)
			if sVal == "ON" {
				relay.On()
			} else {
				relay.Off()
			}
		}
	case server.CmdRelayOn:
		if relay, ok := actors[name]; ok {
			relay.On()
		}
	case server.CmdRelayOff:
		if relay, ok := actors[name]; ok {
			relay.Off()
		}
	case server.CmdGetSensorValue:
		if sensor, ok := sensValues[name]; ok {
			val := fmt.Sprintf("%.4f", sensor)
			msg.ChanReturn <- val
		} else {
			msg.ChanReturn <- "bad"
		}
	default:

	}
}

// OnHandleMessages called when idle to update any messages ect
func OnHandleMessages() {

}

func HandleDevices(sensors map[string]control.ISensor, actors map[string]control.IActor, chnSensor chan control.SensorMessage, sensValues SensorValues) {
	t := time.NewTicker(5000 * time.Millisecond)
	//state := true

	for true {
		select {
		case resvMsg := <-chnSensor:
			name := resvMsg.Name
			fmt.Printf("Recieved from '%s': Value %.3f%s\n", name, resvMsg.Value, sensors[name].GetUnits())
			sensValues[resvMsg.Name] = resvMsg.Value
		case <-t.C:
			OnHandleMessages()
			/*
				for _, act := range actors {
					if state {
						act.On()
					} else {
						act.Off()
					}
					//time.Sleep(time.Millisecond * 250)
				}
				if state {
					state = false
				} else {
					state = true
				}
			*/
		}
	}
}

//RegDevices stores all sensor types that can be used.
type RegDevices map[string]reflect.Type

type SensorValues map[string]float64

var actors map[string]control.IActor
var sensors map[string]control.ISensor

func main() {

	chnSensorValue := make(chan control.SensorMessage)
	svrIn := make(server.SvrChanIn)
	svrOut := make(server.SvrChanOut)

	sensors = make(map[string]control.ISensor)
	actors = make(map[string]control.IActor)
	sensorValues := make(SensorValues)

	regDevices := RegDevices{

		"TempSensor":      reflect.TypeOf(control.TempSensor{}),
		"DummyTempSensor": reflect.TypeOf(control.DummyTempSensor{}),
		"DummyRelay":      reflect.TypeOf(control.DummyRelay{}),
		"SimpleRelay":     reflect.TypeOf(control.SimpleRelay{}),
		"SimpleSSR":       reflect.TypeOf(control.SimpleSSR{}),
	}

	fmt.Println("Starting Controller...")

	logger := control.Logger{}
	logger.Init()
	logger.SetDebug(true)
	logger.Add("default", control.LogLevelAll, os.Stdout)
	availableLinknetAddresses, _ := control.GetActiveNetlinkAddresses(&logger)

	sensorsDefined, senErr := config.DefaultSensorConfig(availableLinknetAddresses)
	if senErr != nil {
		logger.LogMessage("Cannot get sensors configured::%s", senErr.Error())
	}

	for _, sensor := range sensorsDefined {
		if _, ok := regDevices[sensor.Type]; ok {
			t1 := reflect.New(regDevices[sensor.Type]).Interface().(control.ISensor)
			t1.InitSensor(sensor.Name, &logger, toProperties(sensor.Properties), chnSensorValue)
			sensors[sensor.Name] = t1
		}
	}

	for _, sensor := range sensors {
		sensor.OnStart()
		go sensor.Run()
	}

	relaysDefined, relErr := config.DefaultRelayConfig()
	if relErr != nil {
		logger.LogMessage("Cannot get relays configured::%s", relErr.Error())
	}

	for _, actor := range relaysDefined {
		if _, ok := regDevices[actor.Type]; ok {
			t1 := reflect.New(regDevices[actor.Type]).Interface().(control.IActor)
			t1.Init(actor.Name, &logger, toProperties(actor.Properties))
			actors[actor.Name] = t1
		}
	}

	for _, actor := range actors {
		actor.OnStart()
	}

	//	for senMsg := range chnSensorValue {
	//		name := senMsg.Name
	//		fmt.Printf("Recieved from '%s': Value %.3f%s\n", name, senMsg.Value, sensors[name].GetUnits())
	//	}

	go HandleDevices(sensors, actors, chnSensorValue, sensorValues)

	go server.RunWebServer(svrIn, svrOut)

	t := time.NewTicker(5000 * time.Millisecond)

	for true {
		select {
		case in := <-svrOut:
			logger.LogMessage("Got message")
			HandleWebMessage(in, sensorValues)
		case <-t.C:
			logger.LogMessage("tick")
		}
	}
}

func toProperties(propsConfig []config.PropertyConfig) []control.Property {
	props := []control.Property{}

	for _, propCon := range propsConfig {
		prop := toProperty(propCon)
		props = append(props, prop)
	}
	return props
}

func toProperty(propCon config.PropertyConfig) control.Property {
	prop := control.Property{
		Name:     propCon.Name,
		PropType: propCon.Type,
		Hidden:   propCon.Hidden,
		Comment:  propCon.Comment,
		Choice:   propCon.Choice,
		Value:    toValueInterface(propCon.Type, propCon.Value),
	}
	return prop
}

func toValueInterface(sType string, value string) interface{} {
	var i interface{}
	switch sType {
	case "string":
		i = value
	case "uint":
		i, _ = strconv.ParseUint(value, 10, 64)
	default:
		i = value
	}
	return i
}
