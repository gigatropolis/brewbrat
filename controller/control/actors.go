package control

import (
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
	SetState(state DeviceState) (DeviceState, error)
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

func (act *Actor) Init(logger *Logger, properties []Property) error {
	act.Device.Init(logger, properties)

	props := act.GetProperties()
	gpio, ok := props.GetProperty("GPIO")
	if ok {
		act.Pin = gpioreg.ByName(gpio.Value.(string))
		act.LogMessage("Set '%s' to '%s'", gpio.Name, gpio.Value.(string))
	}
	return nil
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

func (act *Actor) GetState() DeviceState {
	return act.state
}

func (act *Actor) SetState(state DeviceState) (DeviceState, error) {
	if state < 0 || state > 100 {
		return 0, nil // TODO return error
	}
	act.state = state

	return act.state, nil
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
