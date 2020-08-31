package control

import (
	"fmt"

	//"periph.io/x/periph/conn/physic"

	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/devices/ds18b20"
	"periph.io/x/periph/experimental/host/netlink"
	"periph.io/x/periph/host"
)

// ISensor defines a Sensor
type ISensor interface {
	IDevice
	GetUnits() string
	OnRead() (float64, error)
	Update(value float64) error
}

// Sensor is base definition for Sensor device.
// your sensor should be defined as a struct with sensor inherited
//    type MySensor struct {
//		Sensor
//		// My Sensor properties here
//	}
type Sensor struct {
	Device
}

// Init called once at sensor creation before OnStart()
func (sen *Sensor) Init(logger *Logger, properties []Property) error {
	sen.Device.Init(logger, properties)
	sen.LogMessage("Init Sensor...")
	return nil
}

func (sen *Sensor) setValue(value float64) {

	select 
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
	Unit       string
}

func (sen *TempSensor) Init(logger *Logger, properties []Property, cnval chan float64) error {
	sen.Sensor.Init(logger, properties)
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
		fmt.Printf("  Failed to open bus: %v", erBus)
		return erBus
	}
	sen.oneBus = oneBus

	// get 1wire address
	addr, erAddr := oneBus.Search(false)
	if erAddr != nil {
		fmt.Printf("  Failed to get 1-wire address(es): %v", erAddr)
		return erAddr
	}

	props := sen.GetProperties()
	sen.Unit = props.InitProperty("Units", "string", "°C", "Units for temperature sensor (default is Celsius)").(string)
	address, ok := props.GetProperty("Address")
	if !ok {
		address.Value = addr[0]
	}

	found := false
	for indx, adrName := range addr {
		fmt.Printf("address(%d)=%d\n", indx, adrName)
		if address.Value.(uint64) == uint64(adrName) {
			found = true
		}
	}

	if !found {
		fmt.Printf("Addess not found for device %s", address)
		return nil // TODO need error codes
	}

	sen.Addresses = append(sen.Addresses, addr...)

	//fmt.Printf("address2=%d", addr[2])
	// init ds18b20
	sensor, erSensor := ds18b20.New(oneBus, addr[0], 12)

	if erSensor != nil {
		fmt.Printf("Failed to get ds18b20 Sensor: %v", erSensor)
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

	if sen.Unit == "°C" {
		return temp.Celsius(), nil
	}

	return temp.Fahrenheit(), nil
}

func (sen *TempSensor) GetUnits() string {
	return sen.Unit
}

func (sen *TempSensor) Update(value float64) error {
	return nil
}

type DummyTempSensor struct {
	Device
	Unit  string
	Value float64
}
