package control

import (
	"time"
)

// Equipment messages
const (
	CmdUpdateDevices = iota + 1
	CmdChangeState
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

func (eq *Equipment) isValidState(state int64) bool {
	if state != EqStateActive &&
		state != EqStateIdle {
		return false
	}
	return true
}

func (eq *Equipment) AddSensor(name string) error {
	eq.Sensors[name] = SensValue{Name: name}
	return nil
}
func (eq *Equipment) AddActor(name string) error {
	eq.Actors[name] = ActValue{Name: name}
	return nil
}

func (eq *Equipment) readMessages() error {
	var err error = nil
	tWait := time.NewTimer(time.Millisecond * 4000)
	select {
	case inMessage := <-eq.in:
		eq.handleMessage(inMessage)
	case <-tWait.C:
	default:
	}
	return err
}

func (eq *Equipment) handleMessage(message EquipMessage) error {

	switch message.Cmd {
	case CmdUpdateDevices:
		for _, sensor := range message.Sensors {
			s := eq.Sensors[sensor.Name]
			s.Value = sensor.Value
			eq.Sensors[sensor.Name] = s
		}
		for _, actor := range message.Actors {
			a := eq.Actors[actor.Name]
			a.State = actor.State
			a.Power = actor.Power
			eq.Actors[actor.Name] = a
		}
	case CmdChangeState:
		if eq.isValidState(message.IntParam1) {
			eq.State = int(message.IntParam1)
		}
	}
	return nil
}

// Run will handle reading in channel and setting values for sensors and actors
func (eq *Equipment) Run() error {

	for true {
		eq.readMessages()
		eq.NextStep()
		//time.Sleep(time.Second * 3)
	}
	return nil
}

func (eq *Equipment) NextStep() error {
	return nil
}

type SimpleRIMM struct {
	Equipment
}

func (rim *SimpleRIMM) InitEquipment(name string, logger *Logger, properties []Property, in <-chan EquipMessage, out chan<- EquipMessage) error {
	rim.Equipment.InitEquipment(name, logger, properties, in, out)
	return nil
}

func (rim *SimpleRIMM) NextStep() error {

	switch rim.State {
	case EqStateActive:
		rim.updateActors()
	}
	return nil
}

func (rim *SimpleRIMM) updateActors() error {

}
