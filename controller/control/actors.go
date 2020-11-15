package control

import (
	"../config"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

type ActorDefinition struct {
	Name       string
	Type       string
	Properties []Property
}

type IActor interface {
	IDevice
	On() error
	Off() error
	SetPower(power int) error
	GetState() DeviceState
	GetPowerLevel() int
}

const (
	StateOff = 0
	StateOn  = 1
)

type Actor struct {
	Device
	state DeviceState
	power int
	GPIO  string
	Pin   gpio.PinIO
}

func (act *Actor) Init(name string, logger *Logger, properties []Property) error {
	act.Device.Init(name, logger, properties)

	props := act.GetProperties()
	gpio, ok := props.GetProperty("GPIO")
	if ok {
		act.Pin = gpioreg.ByName(gpio.Value.(string))
		act.LogMessage("Set '%s' to '%s'", gpio.Name, gpio.Value.(string))
	}
	return nil
}

func (sen *Actor) GetDefaultsConfig() ([]config.PropertyConfig, error) {
	return []config.PropertyConfig{
		{Name: "Name", Type: "string", Hidden: true, Value: "Relay 2", Comment: "relay Name", Choice: ""},
		{Name: "GPIO", Type: "string", Hidden: false, Value: "P1_40", Comment: "GPIO by name", Choice: "", Select: "P1_36,P1_38,P1_40"},
	}, nil

}

func (act *Actor) On() error {
	return nil
}

func (act *Actor) Off() error {
	return nil
}

func (act *Actor) SetPower(power int) error {
	act.power = power
	return nil
}

func (act *Actor) GetPowerLevel() int {
	return act.power
}

func (act *Actor) GetState() DeviceState {
	return act.state
}

type DummyRelay struct {
	Actor
}

func (rel *DummyRelay) OnStart() error {
	rel.Off()
	return nil
}

func (rel *DummyRelay) OnStop() error {
	rel.Off()
	return nil
}

func (rel *DummyRelay) On() error {
	rel.LogMessage("%s ON", rel.Name())
	rel.state = StateOn
	return nil
}

func (rel *DummyRelay) Off() error {

	rel.LogMessage("%s OFF", rel.Name())
	rel.state = StateOff
	return nil
}

type SimpleRelay struct {
	Actor
}

func (rel *SimpleRelay) OnStart() error {

	if _, err := host.Init(); err != nil {
		return err
	}

	rel.Off()

	return nil
}

func (rel *SimpleRelay) OnStop() error {
	rel.Off()
	return nil
}

func (rel *SimpleRelay) On() error {

	err := rel.Pin.Out(gpio.Low)
	if err != nil {
		rel.LogMessage("cannot set value On: %s", err)
	}
	rel.state = StateOn
	return nil
}

func (rel *SimpleRelay) Off() error {

	rel.Pin.Out(gpio.High)
	rel.state = StateOff
	return nil
}

type SimpleSSR struct {
	Actor
}

func (rel *SimpleSSR) OnStart() error {

	if _, err := host.Init(); err != nil {
		return err
	}

	rel.Off()

	return nil
}

func (rel *SimpleSSR) OnStop() error {
	rel.Off()
	return nil
}

func (rel *SimpleSSR) On() error {

	rel.Pin.Out(gpio.High)
	rel.state = StateOn
	return nil
}

func (rel *SimpleSSR) Off() error {

	rel.Pin.Out(gpio.Low)
	rel.state = StateOff
	return nil
}
