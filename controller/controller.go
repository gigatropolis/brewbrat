package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
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
			val := fmt.Sprintf("%.2f", sensor)
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
func HandleWebServer(sensorValues SensorValues, chnWebSvrOut server.SvrChanOut, logger *control.Logger) {
	t := time.NewTicker(5000 * time.Millisecond)

	for true {
		select {
		case in := <-chnWebSvrOut:
			logger.LogMessage("Got message")
			HandleWebMessage(in, sensorValues)
		case <-t.C:
			logger.LogMessage("tick")
		}
	}
}

// HandleDevices  listens on device channels like sensors and equipment to handle incomming messages.
func HandleDevices(sensors map[string]control.ISensor, actors map[string]control.IActor, equipment map[string]control.IEquipment, chnSensor chan control.SensorMessage, chnEquipIn chan control.EquipMessage, chnEquipOut chan control.EquipMessage, sensValues SensorValues) {
	t := time.NewTicker(3000 * time.Millisecond)
	//state := true
	needUpdateSensors := false
	needUpdateActors := false

	for true {
		needUpdateSensors = false
		needUpdateActors = false

		select {
		case resvMsg := <-chnSensor:
			name := resvMsg.Name
			fmt.Printf("Recieved from '%s': Value %.3f%s\n", name, resvMsg.Value, sensors[name].GetUnits())
			sensValues[resvMsg.Name] = resvMsg.Value
			needUpdateSensors = true
		case eqMesg := <-chnEquipOut:
			//fmt.Printf("Recieved from Equipment\n")
			switch eqMesg.Cmd {
			case control.CmdSendNotification:
				sensor, ok := sensors[eqMesg.DeviceName]
				if ok {
					sensor.SendNotification(eqMesg.StrParam1)
				}
			case control.CmdActorOn:
				if relay, ok := actors[eqMesg.DeviceName]; ok {
					relay.On()
					needUpdateActors = true
				}
			case control.CmdActorOff:
				if relay, ok := actors[eqMesg.DeviceName]; ok {
					relay.Off()
					needUpdateActors = true
				}
			}
		case <-t.C:
			OnHandleMessages()
		}

		if needUpdateActors || needUpdateSensors {
			for _, eq := range equipment {
				sens := []control.SensValue{}
				acts := []control.ActValue{}
				if needUpdateSensors {
					for name, senVal := range sensValues {
						//fmt.Printf("senVal '%s' %0.2f\n", name, senVal)
						sens = append(sens, control.SensValue{Name: name, Value: senVal})
					}
				}
				if needUpdateActors {
					for _, act := range actors {
						acts = append(acts, control.ActValue{Name: act.Name(), State: act.GetState(), Power: act.GetPowerLevel()})
					}

				}
				chnEquipIn <- control.EquipMessage{Name: eq.Name(), Cmd: control.CmdUpdateDevices, Sensors: sens, Actors: acts}
			}
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
var buzzers map[string]control.IBuzzer

func main() {

	flgDummy := flag.Bool("dummy", false, "Use dummy configuration")
	flgConfig := flag.String("config", "configuration.xml", "XML configuration file to load")
	flag.Parse()
	if *flgDummy == true {
		fmt.Println("Dummy Configutation used")
	}
	chnSensorValue := make(chan control.SensorMessage, 4)
	svrIn := make(server.SvrChanIn)
	svrOut := make(server.SvrChanOut)
	EqIn := make(chan control.EquipMessage, 4)
	EqOut := make(chan control.EquipMessage, 4)
	chnAlive := make(chan int)

	sensors = make(map[string]control.ISensor)
	actors = make(map[string]control.IActor)
	equipment = make(map[string]control.IEquipment)
	buzzers = make(map[string]control.IBuzzer)

	sensorValues := make(SensorValues)

	regDevices := RegDevices{

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

	var availableLinknetAddresses []uint64
	if !(*flgDummy) {
		availableLinknetAddresses, _ = control.GetActiveNetlinkAddresses(&logger)
	}

	rels := []string{"GPIO21", "GPIO20"}
	ssrs := []string{"GPIO16"}
	defaultConfiguration, conErr := config.DefaultConfiguration(availableLinknetAddresses, rels, ssrs, *flgDummy)
	if conErr != nil {
		logger.LogMessage("Unable to create default configuration::%s", conErr.Error())
	}

	//buf, err := ioutil.ReadFile(*flgConfig)
	configFile, _ := xml.MarshalIndent(defaultConfiguration, "", "   ")
	//fmt.Println(string(configFile))
	ioutil.WriteFile(*flgConfig, configFile, 0644)

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
		go eq.Run()
	}

	for _, buz := range defaultConfiguration.Buzzers {

		if _, ok := regDevices[buz.Type]; ok {
			t1 := reflect.New(regDevices[buz.Type]).Interface().(control.IBuzzer)
			t1.Init(buz.Name, &logger, toProperties(buz.Properties))
			buzzers[buz.Name] = t1
		}
	}

	for _, buzzs := range buzzers {
		buzzs.OnStart()
	}

	buzzers["Main Buzzer"].PlaySound("Main")

	go HandleDevices(sensors, actors, equipment, chnSensorValue, EqIn, EqOut, sensorValues)

	go server.RunWebServer(svrIn, svrOut)

	go HandleWebServer(sensorValues, svrOut, &logger)

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
