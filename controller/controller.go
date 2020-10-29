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

// HandleWebMessage recieves all messages coming from web UI and calls appropriate handlers
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

// OnHandleMessages called when HandleDevices() is idle to do any needed processing.
func OnHandleMessages() {

}

// HandleWebServer recieves all incoming messages from web server
func HandleWebServer(sensorValues SensorValues, chnWebSvrIn server.SvrChanIn, logger *control.Logger) {
	t := time.NewTicker(5000 * time.Millisecond)

	for true {
		select {
		case in := <-chnWebSvrIn:
			logger.LogMessage("Got message")
			HandleWebMessage(in, sensorValues)
		case <-t.C:
			logger.LogMessage("tick")
		}
	}
}

// HandleDevices  listens on device channels like sensors and equipment to handle incomming messages.
func HandleDevices(sensors map[string]control.ISensor, actors map[string]control.IActor, chnSensor chan control.SensorMessage, chnEquipment chan control.EquipMessage, sensValues SensorValues) {
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
		}
	}
}

//RegDevices stores all device types that can be used.
type RegDevices map[string]reflect.Type

// SensorValues stores updated values from all registered sensors
type SensorValues map[string]float64

var actors map[string]control.IActor
var sensors map[string]control.ISensor
var equipment map[string]control.IEquipment
var Buzzers map[string]control.IBuzzer

func main() {

	chnSensorValue := make(chan control.SensorMessage)
	svrIn := make(server.SvrChanIn)
	svrOut := make(server.SvrChanOut)
	EqIn := make(chan control.EquipMessage)
	EqOut := make(chan control.EquipMessage)
	chnAlive := make(chan int)

	sensors = make(map[string]control.ISensor)
	actors = make(map[string]control.IActor)
	equipment = make(map[string]control.IEquipment)
	Buzzers = make(map[string]control.IBuzzer)

	sensorValues := make(SensorValues)

	regDevices := RegDevices{

		"TempSensor":      reflect.TypeOf(control.TempSensor{}),
		"DummyTempSensor": reflect.TypeOf(control.DummyTempSensor{}),
		"DummyRelay":      reflect.TypeOf(control.DummyRelay{}),
		"SimpleRelay":     reflect.TypeOf(control.SimpleRelay{}),
		"SimpleSSR":       reflect.TypeOf(control.SimpleSSR{}),
		"SimpleRIMM":      reflect.TypeOf(control.SimpleRIMM{}),
		"ActiveBuzzer":    reflect.TypeOf(control.ActiveBuzzer{}),
	}

	fmt.Println("Starting Controller...")

	logger := control.Logger{}
	logger.Init()
	logger.SetDebug(true)
	logger.Add("default", control.LogLevelAll, os.Stdout)
	availableLinknetAddresses, _ := control.GetActiveNetlinkAddresses(&logger)

	rels := []string{"GPIO21", "GPIO20"}
	ssrs := []string{"GPIO16"}
	defaultConfiguration, conErr := config.DefaultConfiguration(availableLinknetAddresses, rels, ssrs, false)
	if conErr != nil {
		logger.LogMessage("Unable to create default configuration::%s", conErr.Error())
	}

	for _, sensor := range defaultConfiguration.Sensors {
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

	for _, actor := range defaultConfiguration.Actors {
		if _, ok := regDevices[actor.Type]; ok {
			t1 := reflect.New(regDevices[actor.Type]).Interface().(control.IActor)
			t1.Init(actor.Name, &logger, toProperties(actor.Properties))
			actors[actor.Name] = t1
		}
	}

	for _, actor := range actors {
		actor.OnStart()
	}

	for _, eq := range defaultConfiguration.Equipment {

		if _, ok := regDevices[eq.Type]; ok {
			t1 := reflect.New(regDevices[eq.Type]).Interface().(control.IEquipment)
			t1.InitEquipment(eq.Name, &logger, toProperties(eq.Properties), EqIn, EqOut)
			equipment[eq.Name] = t1
		}
	}

	for _, eq := range equipment {
		eq.OnStart()
	}

	buzzer := reflect.New(regDevices["ActiveBuzzer"]).Interface().(control.IBuzzer)
	buzzer.Init("Main Buzzer", &logger, []control.Property{})
	buzzer.OnStart()
	buzzer.PlaySound("Main")
	Buzzers["Main Buzzer"] = buzzer

	go HandleDevices(sensors, actors, chnSensorValue, EqOut, sensorValues)

	go server.RunWebServer(svrIn, svrOut)

	go HandleWebServer(sensorValues, svrIn, &logger)

	<-chnAlive
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
