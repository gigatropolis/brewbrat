package control

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"time"

	"../config"
	"../www/cmd/server"
)

const (
	RunCmdMode = iota + 1
	ConfigCmdMode
)

type sErr struct {
	msg     string
	errCode int
}

func (er *sErr) String() string {
	return fmt.Sprintf("ERROR (%d) %s", er.errCode, er.msg)
}

// SensorValues stores updated values from all registered sensors
type SensorValues map[string]float64

//RegDevices stores all device types that can be used.
type RegDevices map[string]reflect.Type

type Controller interface {
	InitController(reg *RegDevices, log *Logger, fileName string, isDummyController bool)
	HandleWebMessage(msg server.ServerCommand)
	OnHandleMessages()
	Run()
	OnStart()
}

type Control struct {
	regDevices        *RegDevices
	configuration     *config.BrewController
	logger            *Logger
	isDummyController bool
	configFileName    string
	actors            map[string]IActor
	sensors           map[string]ISensor
	equipment         map[string]IEquipment
	buzzers           map[string]IBuzzer

	chnSensorValue chan SensorMessage
	sensorValues   SensorValues
	svrIn          server.SvrChanIn
	svrOut         server.SvrChanOut
	EqIn           chan EquipMessage
	EqOut          chan EquipMessage
	chnAlive       chan int
}

func (ctrl *Control) InitController(reg *RegDevices, log *Logger, cmdMode int, fileName string, isDummyController bool) error {
	ctrl.regDevices = reg
	ctrl.logger = log
	ctrl.isDummyController = isDummyController
	ctrl.configFileName = fileName

	ctrl.chnSensorValue = make(chan SensorMessage, 4)
	ctrl.svrIn = make(server.SvrChanIn)
	ctrl.svrOut = make(server.SvrChanOut)
	ctrl.EqIn = make(chan EquipMessage, 4)
	ctrl.EqOut = make(chan EquipMessage, 4)
	ctrl.chnAlive = make(chan int)

	ctrl.sensors = make(map[string]ISensor)
	ctrl.actors = make(map[string]IActor)
	ctrl.equipment = make(map[string]IEquipment)
	ctrl.buzzers = make(map[string]IBuzzer)

	ctrl.sensorValues = make(SensorValues)

	ctrl.logger.LogMessage("fileName=%s, mode=%d", fileName, cmdMode)
	if cmdMode == RunCmdMode {
		buf, err := ioutil.ReadFile(fileName)
		if err == nil {

			ctrl.configuration = new(config.BrewController)
			err = xml.Unmarshal(buf, ctrl.configuration)
			if err != nil {
				ctrl.logger.LogError("Unable to parse configuration file: '%s'. Will use default configuration", fileName)
				ctrl.SetDefaultConfiguration()
			}
		} else {
			ctrl.logger.LogError("Unable to read configuration file: '%s'. Will use default configuration", fileName)
			ctrl.SetDefaultConfiguration()
		}
	} else if cmdMode == ConfigCmdMode {
		ctrl.logger.LogMessage("fileName=%s", fileName)
		if fileName == "default" {
			ctrl.configFileName = "configuration.xml"
			ctrl.SetDefaultConfiguration()
		} else {

			ctrl.SetDefaultConfiguration()

			//var availableLinknetAddresses []uint64
			//if !ctrl.isDummyController {
			//	availableLinknetAddresses, _ = GetActiveNetlinkAddresses(ctrl.logger)
			//}

			//rels := []string{"GPIO21", "GPIO20"}
			//ssrs := []string{"GPIO16"}
			//defaultConfiguration, conErr := config.DefaultConfiguration(availableLinknetAddresses, rels, ssrs, ctrl.isDummyController)
			//if conErr != nil {
			//	ctrl.logger.LogMessage("Unable to create default configuration::%s", conErr.Error())
			//}

			//configFile, _ := xml.MarshalIndent(defaultConfiguration, "", "   ")
			//fmt.Println(string(configFile))
			//ioutil.WriteFile(ctrl.configFileName, configFile, 0644)
			//ctrl.configuration = &defaultConfiguration
		}

	} else {
		//ctrl.SetDefaultConfiguration()
		ctrl.logger.LogError("Unknown command mode (%d)", cmdMode)
		return nil
	}
	ctrl.InitializeConfiguration()
	return nil
}

func (ctrl *Control) SetDefaultConfiguration() {

	var availableLinknetAddresses []uint64
	if !ctrl.isDummyController {
		availableLinknetAddresses, _ = GetActiveNetlinkAddresses(ctrl.logger)
	}

	rels := []string{"GPIO21", "GPIO20"}
	ssrs := []string{"GPIO16"}
	defaultConfiguration, conErr := config.DefaultConfiguration(availableLinknetAddresses, rels, ssrs, ctrl.isDummyController)
	if conErr != nil {
		ctrl.logger.LogMessage("Unable to create default configuration::%s", conErr.Error())
	}

	//buf, err := ioutil.ReadFile(*flgConfig)
	configFile, _ := xml.MarshalIndent(defaultConfiguration, "", "   ")
	//fmt.Println(string(configFile))
	ioutil.WriteFile(ctrl.configFileName, configFile, 0644)
	ctrl.configuration = &defaultConfiguration
}

func (ctrl *Control) InitializeConfiguration() {
	for _, sensor := range ctrl.configuration.Sensors {
		if _, ok := (*ctrl.regDevices)[sensor.Type]; ok {
			t1 := reflect.New((*ctrl.regDevices)[sensor.Type]).Interface().(ISensor)
			t1.InitSensor(sensor.Name, ctrl.logger, toProperties(sensor.Properties), ctrl.chnSensorValue)
			ctrl.sensors[sensor.Name] = t1
		}
	}

	for _, actor := range ctrl.configuration.Actors {
		if _, ok := (*ctrl.regDevices)[actor.Type]; ok {
			t1 := reflect.New((*ctrl.regDevices)[actor.Type]).Interface().(IActor)
			t1.Init(actor.Name, ctrl.logger, toProperties(actor.Properties))
			ctrl.actors[actor.Name] = t1
		}
	}

	for _, eq := range ctrl.configuration.Equipment {

		if _, ok := (*ctrl.regDevices)[eq.Type]; ok {
			t1 := reflect.New((*ctrl.regDevices)[eq.Type]).Interface().(IEquipment)
			t1.InitEquipment(eq.Name, ctrl.logger, toProperties(eq.Properties), ctrl.EqIn, ctrl.EqOut)
			ctrl.equipment[eq.Name] = t1
		}
	}

	for _, buz := range ctrl.configuration.Buzzers {

		if _, ok := (*ctrl.regDevices)[buz.Type]; ok {
			t1 := reflect.New((*ctrl.regDevices)[buz.Type]).Interface().(IBuzzer)
			t1.Init(buz.Name, ctrl.logger, toProperties(buz.Properties))
			ctrl.buzzers[buz.Name] = t1
		}
	}

}

func (ctrl *Control) OnStart() {

	for _, sensor := range ctrl.sensors {
		sensor.OnStart()
	}

	for _, actor := range ctrl.actors {
		actor.OnStart()
	}

	for _, eq := range ctrl.equipment {
		eq.OnStart()
	}

	for _, buzzs := range ctrl.buzzers {
		buzzs.OnStart()
	}

}

func (ctrl *Control) Run() {

	for _, sensor := range ctrl.sensors {
		go sensor.Run()
	}

	for _, eq := range ctrl.equipment {
		go eq.Run()
	}

	ctrl.buzzers["Main Buzzer"].PlaySound("Main")

	go ctrl.HandleDevices()

	ctrl.logger.LogMessage("Web Server running at 127.0.0.1:8090")
	go server.RunWebServer(ctrl.svrIn, ctrl.svrOut)

	go ctrl.HandleWebServer()

	<-ctrl.chnAlive

}

// HandleWebMessage recieves all messages coming from web UI and calls appropriate handlers
func (ctrl *Control) HandleWebMessage(msg server.ServerCommand) {

	//name := strings.ReplaceAll(msg.DeviceName, "_", " ")
	name := msg.DeviceName
	switch msg.Cmd {
	case server.CmdSetRelay:
		relay, ok := ctrl.actors[name]
		if ok {
			sVal := string(msg.Value)
			if sVal == "ON" {
				relay.On()
			} else {
				relay.Off()
			}
			state := relay.GetState()
			if state == StateOn {
				msg.ChanReturn <- "ON"
			} else {
				msg.ChanReturn <- "OFF"
			}
		} else {
			msg.ChanReturn <- "ack"
		}
	case server.CmdRelayOn:
		if relay, ok := ctrl.actors[name]; ok {
			relay.On()
		}
		msg.ChanReturn <- "ack"
	case server.CmdRelayOff:
		if relay, ok := ctrl.actors[name]; ok {
			relay.Off()
		}
		msg.ChanReturn <- "ack"
	case server.CmdGetSensorValue:
		if sensor, ok := ctrl.sensorValues[name]; ok {
			val := fmt.Sprintf("%.2f", sensor)
			msg.ChanReturn <- val
		} else {
			msg.ChanReturn <- "bad"
		}
	case server.CmdGetActorValue:
		if relay, ok := ctrl.actors[name]; ok {
			state := relay.GetState()
			if state == StateOn {
				//ctrl.logger.LogMessage("server.CmdGetActorValue %s ON", name)
				msg.ChanReturn <- "ON"
			} else {
				ctrl.logger.LogMessage("server.CmdGetActorValue %s OFF", name)
				//msg.ChanReturn <- "OFF"
			}
		} else {
			msg.ChanReturn <- "bad"
		}
	case server.CmdGetSetpointValue:
		if eq, ok := ctrl.equipment[name]; ok {
			setpoint, err := eq.GetSetpoint()
			if err != nil {
				msg.ChanReturn <- "bad"
			} else {
				val := fmt.Sprintf("%0.2f", setpoint)
				msg.ChanReturn <- val
			}
		} else {
			msg.ChanReturn <- "bad"
		}
	default:
		msg.ChanReturn <- "Unknown"
	}
}

// OnHandleMessages called when HandleDevices() is idle to do any needed processing.
func (ctrl *Control) OnHandleMessages() {

}

// HandleWebServer recieves all incoming messages from web server
func (ctrl *Control) HandleWebServer() {
	t := time.NewTicker(5000 * time.Millisecond)
	tickCount := 0
	for true {
		select {
		case in := <-ctrl.svrOut:
			//ctrl.logger.LogMessage("Got message")
			ctrl.HandleWebMessage(in)
		case <-t.C:
			if tickCount > 10 {
				ctrl.logger.LogMessage("tick")
				tickCount = 0
			}
		}
	}
	tickCount++
}

// HandleDevices  listens on device channels like sensors and equipment to handle incomming messages.
func (ctrl *Control) HandleDevices() {
	t := time.NewTicker(3000 * time.Millisecond)
	//state := true
	needUpdateSensors := false
	needUpdateActors := false

	for true {
		needUpdateSensors = false
		needUpdateActors = false

		select {
		case resvMsg := <-ctrl.chnSensorValue:
			//name := resvMsg.Name
			//fmt.Println("Recieved from '%s': Value %.3f\n", name, resvMsg.Value)
			ctrl.sensorValues[resvMsg.Name] = resvMsg.Value
			needUpdateSensors = true
		case eqMesg := <-ctrl.EqOut:
			//fmt.Printf("Recieved from Equipment (%d) '%s' param(%s)\n", eqMesg.Cmd, eqMesg.DeviceName, eqMesg.StrParam1)
			switch eqMesg.Cmd {
			case CmdSendNotification:
				sensor, ok := ctrl.sensors[eqMesg.DeviceName]
				if ok {
					sensor.SendNotification(eqMesg.StrParam1)
				}
			case CmdActorOn:
				if relay, ok := ctrl.actors[eqMesg.DeviceName]; ok {
					relay.On()
					needUpdateActors = true
				}
			case CmdActorOff:
				if relay, ok := ctrl.actors[eqMesg.DeviceName]; ok {
					relay.Off()
					needUpdateActors = true
				}
			}
		case <-t.C:
			ctrl.OnHandleMessages()
		}

		if needUpdateActors || needUpdateSensors {
			for _, eq := range ctrl.equipment {
				sens := []SensValue{}
				acts := []ActValue{}
				if needUpdateSensors {
					for name, senVal := range ctrl.sensorValues {
						//fmt.Printf("senVal '%s' %0.2f\n", name, senVal)
						sens = append(sens, SensValue{Name: name, Value: senVal})
					}
				}
				if needUpdateActors {
					for _, act := range ctrl.actors {
						acts = append(acts, ActValue{Name: act.Name(), State: act.GetState(), Power: act.GetPowerLevel()})
					}

				}
				ctrl.EqIn <- EquipMessage{Name: eq.Name(), Cmd: CmdUpdateDevices, Sensors: sens, Actors: acts}
			}
		}

	}
}

func toProperties(propsConfig []config.PropertyConfig) []Property {
	props := []Property{}

	for _, propCon := range propsConfig {
		prop := toProperty(propCon)
		props = append(props, prop)
	}
	return props
}

func toProperty(propCon config.PropertyConfig) Property {
	prop := Property{
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
	case "int":
		i, _ = strconv.ParseInt(value, 10, 64)
	case "float":
		f, _ := strconv.ParseFloat(value, 64)
		// fmt.Println("toValueInterface (float) =", f)
		return f
	default:
		i = value
	}
	return i
}
