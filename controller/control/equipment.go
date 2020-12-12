package control

import (
	"time"

	"../config"
)

// Equipment messages
const (
	CmdUpdateDevices = iota + 1
	CmdUpdateSensor
	CmdChangeState
	CmdActorOn
	CmdActorOff
	CmdActorChange
	CmdSendNotification
)

const (
	EqStateIdle = iota + 1
	EqStateActive
)

const (
	EqModePIDControl = iota + 1
	EqModeHistorisis
)

type SensValue struct {
	Name  string
	Value float64
}
type ActValue struct {
	Name  string
	State DeviceState
	Power int
}

type EquipMessage struct {
	Name       string
	DeviceName string
	Cmd        int
	FltParam1  float64
	FltParam2  float64
	IntParam1  int64
	IntParam2  int64
	StrParam1  string
	StrParam2  string
	Sensors    []SensValue
	Actors     []ActValue
}

type IEquipment interface {
	IDevice
	InitEquipment(name string, logger *Logger, properties []Property, in <-chan EquipMessage, out chan<- EquipMessage) error
	AddSensor(name string) error
	AddActor(name string) error
	Run() error
	NextStep() error
}

type Equipment struct {
	Device
	State   int
	Mode    int
	Sensors map[string]SensValue
	Actors  map[string]ActValue
	in      <-chan EquipMessage
	out     chan<- EquipMessage
}

// InitEquipment does that
func (eq *Equipment) InitEquipment(name string, logger *Logger, properties []Property, in <-chan EquipMessage, out chan<- EquipMessage) error {
	eq.Device.Init(name, logger, properties)

	eq.in = in
	eq.out = out
	eq.Sensors = make(map[string]SensValue)
	eq.Actors = make(map[string]ActValue)

	props := eq.GetProperties()
	mode := props.InitProperty("Control Mode", "string", "Historisis", "Control mode for equipment").(string)

	switch mode {
	case "Historisis":
		eq.Mode = EqModeHistorisis
	case "PID":
		eq.Mode = EqModePIDControl
	default:
		eq.Mode = EqModeHistorisis
	}

	return nil
}

func (eq *Equipment) GetDefaultsConfig() ([]config.PropertyConfig, error) {
	return []config.PropertyConfig{
		{Name: "Temperature Sensor", Type: "string", Hidden: false, Value: "Dummy Temp 1", Comment: "Sensor Name", Choice: ""},
		{Name: "Units", Type: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""},
		{Name: "Pump", Type: "string", Hidden: false, Value: "Relay 1", Comment: "Units for Sensor", Choice: ""},
		{Name: "Circulator", Type: "string", Hidden: false, Value: "Relay 2", Comment: "Units for Sensor", Choice: ""},
		{Name: "Heater", Type: "string", Hidden: false, Value: "SSR 1", Comment: "Units for Sensor", Choice: ""},
	}, nil

}

func (eq *Equipment) isValidState(state int64) bool {
	if state != EqStateActive &&
		state != EqStateIdle {
		return false
	}
	return true
}

func (eq *Equipment) AddSensor(name string) error {
	eq.LogMessage("eq.AddSensor %s", name)
	eq.Sensors[name] = SensValue{Name: name}
	return nil
}
func (eq *Equipment) AddActor(name string) error {
	eq.LogMessage("eq.AddActor %s", name)
	eq.Actors[name] = ActValue{Name: name}
	return nil
}

func (eq *Equipment) getActorControl(name string) string {
	if name == "Relay 1" {
		return "Temp Sensor 1"
	} else if name == "Relay 2" {
		return "Temp Sensor 2"
	} else if name == "Relay 3" {
		return "Temp Sensor 3"
	}

	return "Relay 1"
}

func (eq *Equipment) readMessages() error {
	var err error = nil
	tWait := time.NewTimer(time.Millisecond * 4000)
	readMessages := true
	for readMessages {
		select {
		case inMessage := <-eq.in:
			eq.handleMessage(inMessage)
		case <-tWait.C:
			readMessages = false
		}
	}
	return err
}

func (eq *Equipment) handleMessage(message EquipMessage) error {
	//eq.LogMessage("eq.handleMessage %d", message.Cmd)
	switch message.Cmd {
	case CmdUpdateDevices:
		for _, sensor := range message.Sensors {
			//eq.LogMessage("start CmdUpdateDevices: eq.handleMessage %s", sensor.Name)
			s, ok := eq.Sensors[sensor.Name]
			if ok {
				//eq.LogMessage("sensor.Name: eq.handleMessage %s", sensor.Name)
				s.Value = sensor.Value
				eq.Sensors[sensor.Name] = s
			}
		}
		for _, actor := range message.Actors {
			//eq.LogMessage("start CmdUpdateDevices::eq.handleMessage '%s'", actor.Name)
			a, ok := eq.Actors[actor.Name]
			if ok {
				if a.State != actor.State &&
					eq.IsDummyDevice() {
					//eq.LogMessage("start CmdUpdateDevices: eq.handleMessage '%s' Handled", actor.Name)
					cmd := "OFF"
					if actor.State == StateOn {
						cmd = "ON"
					}
					//eq.LogMessage("Actor out %s", actor.Name)
					name := eq.getActorControl(actor.Name)

					eq.out <- EquipMessage{DeviceName: name, Cmd: CmdSendNotification, StrParam1: cmd}
				}
				a.State = actor.State
				a.Power = actor.Power
				eq.Actors[actor.Name] = a
			}
		}
	case CmdChangeState:
		if eq.isValidState(message.IntParam1) {
			eq.State = int(message.IntParam1)
		}
	}
	return nil
}

// OnStart called once when device first started up. Called after Init()
func (eq *Equipment) OnStart() error {

	eq.State = EqStateActive
	return nil
}

type SimpleRIMM struct {
	Equipment
	PowerOn       int
	PowerOff      int
	TempProbeName string
	HeaterName    string
	PumpName      string
	AgitatorName  string
}

func (rim *SimpleRIMM) InitEquipment(name string, logger *Logger, properties []Property, in <-chan EquipMessage, out chan<- EquipMessage) error {
	rim.Equipment.InitEquipment(name, logger, properties, in, out)

	props := rim.GetProperties()
	rim.PowerOn = props.InitProperty("Power On", "int", 147, "Power goes on if temperature drops below this value").(int)
	rim.PowerOff = props.InitProperty("Power Off", "int", 150, "Power goes Off if temperature goes above this value").(int)
	rim.TempProbeName = props.InitProperty("Temperature Sensor", "string", "Temp Sensor 1", "Name of Temperature Sensor").(string)
	rim.HeaterName = props.InitProperty("Pump Name", "string", "Relay 1", "Name of actor used to control Heater").(string)
	rim.PumpName = props.InitProperty("Heater Name", "string", "Relay 2", "Name of actor used to control Pump").(string)
	rim.AgitatorName = props.InitProperty("Agitator Name", "string", "Relay 3", "Name of actor used to for agitation").(string)

	rim.AddSensor(rim.TempProbeName)
	rim.AddActor(rim.HeaterName)
	rim.AddActor(rim.PumpName)
	rim.AddActor(rim.AgitatorName)
	return nil
}

// Run will handle reading in channel and setting values for sensors and actors
func (rim *SimpleRIMM) Run() error {

	for true {
		rim.readMessages()
		rim.NextStep()
		//time.Sleep(time.Second * 3)
	}
	return nil
}

func (rim *SimpleRIMM) NextStep() error {
	rim.LogMessage("rim.NextStep")

	switch rim.State {
	case EqStateActive:
		rim.updateActors()
	}
	return nil
}

func (rim *SimpleRIMM) updateActors() error {
	rim.LogMessage("rim.updateActors")

	var err error = nil
	switch rim.Mode {
	case EqModeHistorisis:
		err = rim.updateHistorisis()
	case EqModePIDControl:
		err = rim.updatePID()
	}
	return err
}

func (rim *SimpleRIMM) updateHistorisis() error {

	temp, ok := rim.Sensors[rim.TempProbeName]
	//rim.LogMessage("temp.Value %0.2f", temp.Value)
	if !ok {
		return nil
	}
	if int(temp.Value) > rim.PowerOff {
		if act, ok := rim.Actors[rim.HeaterName]; ok {
			if act.State != StateOff {
				rim.out <- EquipMessage{DeviceName: rim.HeaterName, Cmd: CmdActorOff}
			}
		}
	}
	if int(temp.Value) < rim.PowerOn {
		if act, ok := rim.Actors[rim.HeaterName]; ok {
			if act.State != StateOn {
				rim.out <- EquipMessage{DeviceName: rim.HeaterName, Cmd: CmdActorOn}
			}
		}
	}
	return nil
}

func (rim *SimpleRIMM) updatePID() error {
	var err error = nil
	return err
}
