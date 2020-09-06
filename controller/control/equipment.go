package control

// Equipment messages
const (
	CmdUpdateDevices = iota + 1
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
	Sensors    []SensValue
	Actors     []ActValue
}

type IEquipment interface {
	IDevice
	InitEquipment(logger *Logger, properties []Property, in <-chan EquipMessage, out chan<- EquipMessage) error
	Run() error
}

type Equipment struct {
	Device
	Sensors map[string]SensValue
	Actors  map[string]ActValue
	in      <-chan EquipMessage
	out     chan<- EquipMessage
}

func (eq *Equipment) InitEquipment(logger *Logger, properties []Property, in <-chan EquipMessage, out chan<- EquipMessage) error {
	eq.Device.Init(logger, properties)

	eq.in = in
	eq.out = out
	eq.Sensors = make(map[string]SensValue)
	eq.Actors = make(map[string]ActValue)
	return nil
}

func (eq *Equipment) readMessage() error {
	var err error = nil
	select {
	case inMessage := <-eq.in:
		eq.handleMessage(inMessage)
	default:
	}
	return err
}

func (eq *Equipment) handleMessage(message EquipMessage) error {

	switch message.Cmd {
	case CmdUpdateDevices:
		for _, sensor := range message.Sensors {
			s := eq.Sensors[sensor.Name]
			s.Name = sensor.Name
			s.Value = sensor.Value
		}
		for _, actor := range message.Actors {
			a := eq.Actors[actor.Name]
			a.Name = actor.Name
			a.State = actor.State
			a.Power = actor.Power
		}
	}
	return nil
}

// Run will handle reading in channel and setting values for sensors and actors
func (eq *Equipment) Run() error {
	eq.readMessage()
	return nil
}

type SimpleRIMM struct {
	Equipment
}
