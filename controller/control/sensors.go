package control

import (
	"time"
	//"periph.io/x/periph/conn/physic"

	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/devices/ds18b20"
	"periph.io/x/periph/experimental/host/netlink"
	"periph.io/x/periph/host"
)

type SensorMessage struct {
	Name  string
	Value float64
}

// ISensor defines a Sensor
type ISensor interface {
	IDevice
	InitSensor(logger *Logger, properties []Property, cnval chan<- SensorMessage) error
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
func (sen *Sensor) InitSensor(logger *Logger, properties []Property, cnval chan<- SensorMessage) error {
	sen.Device.Init(logger, properties)
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

// Run is main loop for Sensor that will be launched by Brebrat in a seperate go routine.
// This method calls OnRead() in the main loop.
// User doesn't need to override Run() methos but at least override OnRead() to get sensor value.
// Override this method to change default behavior
func (sen *Sensor) Run() error {

	active := true
	for active {
		value, err := sen.OnRead()
		if err != nil {
			sen.LogMessage("can't read sensor")
			active = false
		} else {
			err := sen.SetValue(value)
			if err != nil {
				sen.LogMessage("can't set sensor value")
			}
			//sen.LogMessage("Sensor value = %.3f%s", value, t1.GetUnits())
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

// InitSensor must initialize 1-wire host and call base init
func (sen *TempSensor) InitSensor(logger *Logger, properties []Property, cnval chan<- SensorMessage) error {
	sen.Sensor.InitSensor(logger, properties, cnval)
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
	if !ok {
		address.Value = addr[0]
	}

	found := false
	for indx, adrName := range addr {
		sen.LogMessage("address(%d)=%d\n", indx, adrName)
		if address.Value.(uint64) == uint64(adrName) {
			found = true
		}
	}

	if !found {
		sen.LogMessage("Addess not found for device %s", address)
		return nil // TODO need error codes
	}

	sen.Addresses = append(sen.Addresses, addr...)

	//fmt.Printf("address2=%d", addr[2])
	// init ds18b20
	sensor, erSensor := ds18b20.New(oneBus, addr[0], 12)

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

func (sen *TempSensor) Update(value float64) error {
	return nil
}

// DummyTempSensor moves temperatures up and down
type DummyTempSensor struct {
	Sensor
	MaxTemp float64
	temp    float64
	change  float64
	offset  float64
	cnt     int
}

// InitSensor must initialize 1-wire host and call base init
func (sen *DummyTempSensor) InitSensor(logger *Logger, properties []Property, cnval chan<- SensorMessage) error {
	sen.Sensor.InitSensor(logger, properties, cnval)
	sen.LogMessage("init DummyTempSensor...")
	return nil
}

// OnStart setup to start running
func (sen *DummyTempSensor) OnStart() error {
	if sen.GetUnits() == "°C" {
		sen.temp = 50
		sen.change = 0.01
		sen.offset = 0.01
	} else {
		sen.temp = 85
		sen.change = 0.01
		sen.offset = 0.01
	}

	sen.cnt = 20

	return nil
}
