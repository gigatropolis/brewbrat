package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"./www/cmd/server"

	"./config"
	"./control"
)

func HandleWebMessage(msg server.ServerCommand) {
	switch msg.Cmd {
	case server.CmdSetRelay:
		name := msg.DeviceName
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
	case server.CmdRelayOff:
	default:

	}
}
func HandleDevices(sensors map[string]control.ISensor, actors map[string]control.IActor, chnSensor chan control.SensorMessage) {
	t := time.NewTicker(5000 * time.Millisecond)
	state := true

	for true {
		select {
		case resvMsg := <-chnSensor:
			name := resvMsg.Name
			fmt.Printf("Recieved from '%s': Value %.3f%s\n", name, resvMsg.Value, sensors[name].GetUnits())
		case <-t.C:
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
		}
	}
}

//RegDevices stores all sensor types that can be used.
type RegDevices map[string]reflect.Type

var actors map[string]control.IActor
var sensors map[string]control.ISensor

func main() {

	chnSensorValue := make(chan control.SensorMessage)
	svrIn := make(server.SvrChanIn)
	svrOut := make(server.SvrChanOut)

	sensors = make(map[string]control.ISensor)
	actors = make(map[string]control.IActor)

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

	sensorsDefined, senErr := config.DefaultSensorConfig()
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

	go HandleDevices(sensors, actors, chnSensorValue)

	go server.RunWebServer(svrIn, svrOut)

	t := time.NewTicker(5000 * time.Millisecond)

	for true {
		select {
		case in := <-svrOut:
			logger.LogMessage("Got message")
			HandleWebMessage(in)
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
