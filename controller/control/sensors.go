package control

import (
	"time"
	//"periph.io/x/periph/conn/physic"

	"../config"
	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/devices/ds18b20"
	"periph.io/x/periph/experimental/host/netlink"
	"periph.io/x/periph/host"
)

type SensorDefinition struct {
	Name       string
	Type       string
	Properties []Property
}

type SensorMessage struct {
	Name  string
	Value float64
}

// ISensor defines a Sensor
type ISensor interface {
	IDevice
	InitSensor(name string, logger *Logger, properties []Property, cnval chan<- SensorMessage) error
	GetUnits() string
	OnRead() (float64, error)
	SetValue(float64) error
	Run() error
}

// Sensor is base definition for Sensor device.
// your sensor should be defined as a struct with sensor inherited
//    type MySensor struct {
//		Sensor
//		// My Sensor properties here
//	}
type Sensor struct {
	Device
	chnValue chan<- SensorMessage
	Unit     string
}

// InitSensor called once at sensor creation before OnStart()
func (sen *Sensor) InitSensor(name string, logger *Logger, properties []Property, cnval chan<- SensorMessage) error {
	sen.Device.Init(name, logger, properties)
	sen.LogMessage("Init Sensor...")
	sen.chnValue = cnval
	props := sen.GetProperties()
	sen.Unit = props.InitProperty("Units", "string", "°C", "Units for temperature sensor (default is Celsius)").(string)
	return nil
}

func (sen *Sensor) GetUnits() string {
	return sen.Unit
}

func (sen *Sensor) OnRead() (float64, error) {
	return 99.99, nil
}

func (sen *Sensor) SetValue(value float64) error {

	select {
	case sen.chnValue <- SensorMessage{Name: sen.Name(), Value: value}:
	case <-time.After(time.Millisecond * 5000):
	}
	return nil
}

// Run is main loop for Sensor that will be launched by Brewbrat in a seperate go routine.
// This method calls OnRead() in the main loop.
// User doesn't need to override Run() methos but at least override OnRead() to get sensor value.
// Override this method to change default behavior
func (sen *Sensor) startRun(fRead func() (float64, error)) error {
	sen.LogMessage("Start Run %s", sen.Name())
	active := true
	for active {
		value, err := fRead()
		if err != nil {
			sen.LogMessage("can't read sensor")
			active = false
		} else {
			err := sen.SetValue(value)
			if err != nil {
				sen.LogMessage("can't set sensor value")
			}
			//sen.LogMessage("Sensor value = %.3f%s", value, sen.GetUnits())
		}
		time.Sleep(time.Second * 3)
	}
	return nil
}

// TempSensor is a 1-Wire DS18B20 temperature sensor
// Uses netlink bus for communication
// each temp sensor will have unique UINT64 Address
type TempSensor struct {
	Sensor
	oneBus     *netlink.OneWire
	Addresses  []onewire.Address
	Address    string
	RealDevice *ds18b20.Dev
}

func (sen *TempSensor) GetDefaultsConfig() ([]config.PropertyConfig, error) {
	return []config.PropertyConfig{
		{Name: "Name", Type: "string", Hidden: true, Value: "temp Sensor 1", Comment: "Sensor Name", Choice: ""},
		{Name: "Address", Type: "uint", Hidden: false, Value: "7205759448148251176", Comment: "1-Wire sensor address", Choice: ""},
		{Name: "Units", Type: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""},
	}, nil

}

// InitSensor must initialize 1-wire host and call base init
func (sen *TempSensor) InitSensor(name string, logger *Logger, properties []Property, cnval chan<- SensorMessage) error {
	sen.Sensor.InitSensor(name, logger, properties, cnval)
	sen.LogMessage("init TempSensor...")

	if _, err := host.Init(); err != nil {
		return err
	}
	return nil
}

func (sen *TempSensor) OnStart() error {
	// get 1wire bus
	oneBus, erBus := netlink.New(001)
	if erBus != nil {
		sen.LogMessage("Failed to open bus: %v", erBus)
		return erBus
	}
	sen.oneBus = oneBus

	// get 1wire address
	addr, erAddr := oneBus.Search(false)
	if erAddr != nil {
		sen.LogMessage("Failed to get 1-wire address(es): %v", erAddr)
		return erAddr
	}

	props := sen.GetProperties()
	address, ok := props.GetProperty("Address")
	sen.LogMessage("%s = %d", sen.Name(), address.Value.(uint64))
	if !ok {
		sen.LogMessage("Address Property Not found")
		address.Value = addr[0]
	}

	found := false
	for indx, adrName := range addr {
		sen.LogMessage("address(%d)=%d\n", indx, adrName)
		if address.Value.(uint64) == uint64(adrName) {
			address.Value = onewire.Address(adrName)
			found = true
			break
		}
	}

	if !found {
		sen.LogMessage("Addess not found for device %s", address)
		return nil // TODO need error codes
	}

	sen.Addresses = append(sen.Addresses, addr...)

	//fmt.Printf("address2=%d", addr[2])
	// init ds18b20
	sensor, erSensor := ds18b20.New(oneBus, address.Value.(onewire.Address), 12)

	if erSensor != nil {
		sen.LogMessage("Failed to get ds18b20 Sensor: %v", erSensor)
		return erSensor
	}
	sen.RealDevice = sensor
	return nil
}

func (sen *TempSensor) OnStop() error {
	sen.oneBus.Close()
	return nil
}

// OnRead called in default loop of Run() method.
// Use this method to return get returned from sensor
func (sen *TempSensor) OnRead() (float64, error) {

	ds18b20.ConvertAll(sen.oneBus, 12)
	temp, _ := sen.RealDevice.LastTemp()

	//fmt.Printf("%s %.4f°F\n", temp, temp.Fahrenheit())
	//time.Sleep(5 * time.Second)

	if sen.GetUnits() == "°C" {
		return temp.Celsius(), nil
	}

	return temp.Fahrenheit(), nil
}

// Run can call sen.startRun(sen.OnRead) for default behavior
// Use value returnd from sen.OnRead()
func (sen *TempSensor) Run() error {
	sen.startRun(sen.OnRead)
	return nil
}

// DummyTempSensor moves temperatures up and down
type DummyTempSensor struct {
	Sensor
	MaxTemp   float64
	minTemp   float64
	prevState string
	state     string
	temp      float64
	change    float64
	offset    float64
	cnt       int
	direction float64
}

// InitSensor must initialize 1-wire host and call base init
func (sen *DummyTempSensor) InitSensor(name string, logger *Logger, properties []Property, cnval chan<- SensorMessage) error {
	sen.Sensor.InitSensor(name, logger, properties, cnval)
	sen.LogMessage("init DummyTempSensor (%s)...", sen.GetUnits())
	return nil
}

func (sen *DummyTempSensor) GetDefaultsConfig() ([]config.PropertyConfig, error) {
	return []config.PropertyConfig{
		{Name: "Name", Type: "string", Hidden: false, Value: "Dummy Temp 1", Comment: "Sensor Name", Choice: ""},
		{Name: "Units", Type: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""},
	}, nil

}

func (sen *DummyTempSensor) SendNotification(notify string) error {
	sen.LogMessage("%s: set sen.state = %s", sen.Name(), notify)
	sen.state = notify
	return nil
}

// OnStart setup to start running
func (sen *DummyTempSensor) OnStart() error {
	if sen.GetUnits() == "°C" {
		sen.temp = 50
		sen.MaxTemp = 100
	} else {
		sen.temp = 110
		sen.MaxTemp = 212
	}
	sen.change = 0.1
	sen.offset = 0.1

	sen.cnt = 0
	sen.direction = -1
	sen.minTemp = sen.temp - 10

	sen.prevState = "OFF"
	sen.state = "OFF"
	return nil
}

func (sen *DummyTempSensor) OnRead() (float64, error) {

	sen.temp += sen.change * sen.direction
	if sen.cnt > 3 {
		sen.change += (sen.offset)
		sen.cnt = 0
	}
	temp := sen.temp

	if temp > sen.MaxTemp+1 {
		temp = sen.MaxTemp
	} else if temp < sen.minTemp {
		temp = sen.minTemp
	} else if sen.prevState != sen.state {
		sen.LogMessage("%s != %s", sen.prevState, sen.state)
		if sen.state == "ON" {
			sen.LogMessage("%s state to '%s'", sen.Name(), sen.state)
			sen.direction = 1
		} else {
			sen.LogMessage("%s state to '%s'", sen.Name(), sen.state)
			sen.direction = -1
		}
		sen.prevState = sen.state
		sen.change = 0.1
	}
	sen.cnt++
	sen.temp = temp

	return temp, nil
}

// Run can call sen.startRun(sen.OnRead) for default behavior
// Use value returnd from sen.OnRead()
func (sen *DummyTempSensor) Run() error {
	sen.startRun(sen.OnRead)
	return nil
}

func GetActiveNetlinkAddresses(logger *Logger) ([]uint64, error) {

	addresses := []uint64{}

	if _, err := host.Init(); err != nil {
		return nil, err
	}

	oneBus, erBus := netlink.New(001)
	if erBus != nil {
		logger.LogMessage("Failed to open bus: %v", erBus)
		return addresses, erBus
	}

	// get 1wire address
	addr, erAddr := oneBus.Search(false)
	if erAddr != nil {
		logger.LogMessage("Failed to get 1-wire address(es): %v", erAddr)
		return addresses, erAddr
	}

	for indx, adrName := range addr {
		logger.LogMessage("address(%d)=%d\n", indx, adrName)
		addresses = append(addresses, uint64(adrName))
	}
	return addresses, nil
}
