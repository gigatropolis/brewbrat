package control

import (
	"time"

	"../config"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

// SoundBit id piece of sound that has On and Off each Milliseconds
type SoundBit struct {
	Level int
	On    int
	Off   int
}

type IBuzzer interface {
	IDevice
	On() error
	Off() error
	PlaySound(name string) error
	PlaySoundBit(soundBits []SoundBit) error
}

type Buzzer struct {
	Device
	Sounds map[string][]SoundBit
	GPIO   string
	Pin    gpio.PinIO
}

func (buz *Buzzer) Init(name string, logger *Logger, properties []Property) error {
	buz.Device.Init(name, logger, properties)

	props := buz.GetProperties()
	buz.GPIO = props.InitProperty("GPIO", "string", "GPIO18", "GPIO used to control buzzer").(string)
	buz.Pin = gpioreg.ByName(buz.GPIO)
	buz.LogMessage("Set buzzer %s '%s'", name, buz.GPIO)

	buz.Sounds = make(map[string][]SoundBit)
	buz.Sounds["Main"] = []SoundBit{
		{100, 200, 20},
		{100, 200, 20},
		{100, 200, 20},
	}
	return nil
}

func (buz *Buzzer) OnStart() error {

	if _, err := host.Init(); err != nil {
		return err
	}

	buz.Off()
	return nil
}

func (buz *Buzzer) OnStop() error {
	buz.Off()
	return nil
}

func (buz *Buzzer) On() error {

	err := buz.Pin.Out(gpio.High)
	if err != nil {
		buz.LogMessage("cannot set value On: %s", err)
	}
	return nil
}

func (buz *Buzzer) Off() error {

	buz.Pin.Out(gpio.Low)
	return nil
}

type ActiveBuzzer struct {
	Buzzer
}

func (buz *ActiveBuzzer) GetDefaultsConfig() ([]config.PropertyConfig, error) {
	return []config.PropertyConfig{
		{Name: "Name", Type: "string", Hidden: true, Value: "Main Buzzer", Comment: "Buzzer Name", Choice: ""},
		{Name: "GPIO", Type: "string", Hidden: false, Value: "GPIO18", Comment: "GPIO by name", Choice: "", Select: "GPIO18,GPIO22"},
	}, nil

}

func (buz *ActiveBuzzer) PlaySound(name string) error {

	if sound, ok := buz.Sounds[name]; ok {
		buz.PlaySoundBit(sound)
	}

	return nil
}

func (buz *ActiveBuzzer) PlaySoundBit(soundBits []SoundBit) error {
	buz.LogMessage("Play Sound")
	for _, bit := range soundBits {
		buz.On()
		time.Sleep(time.Duration(bit.On) * time.Millisecond)
		buz.Off()
		time.Sleep(time.Duration(bit.Off) * time.Millisecond)
	}

	buz.Off()

	return nil
}

type DummyBuzzer struct {
	Buzzer
}

func (buz *DummyBuzzer) Init(name string, logger *Logger, properties []Property) error {
	properties = append(properties, CreateDummyProp())

	buz.Device.Init(name, logger, properties)
	return nil
}

func (buz *DummyBuzzer) On() error {
	return nil
}

func (buz *DummyBuzzer) Off() error {
	return nil
}

func (buz *DummyBuzzer) PlaySound(name string) error {
	return nil
}

func (buz *DummyBuzzer) PlaySoundBit(soundBits []SoundBit) error {
	buz.LogMessage("Play Sound it")
	return nil
}
