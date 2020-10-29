package control

import (
	"../config"
)

type IDevice interface {
	Init(name string, logger *Logger, properties []Property) error
	OnStart() error
	OnStop() error
	Name() string
	String() string
	IsDummyDevice() bool
	SendNotification(notify string) error
	GetDefaultsConfig() ([]config.PropertyConfig, error)
	LogMessage(pattern string, args ...interface{}) error
	LogWarning(pattern string, args ...interface{}) error
	LogError(pattern string, args ...interface{}) error
	LogDebug(pattern string, args ...interface{}) error
}

type DeviceState int

type DeviceDefinition struct {
	Name       string
	DevType    string
	DevClass   string
	Properties []Property
}

type Device struct {
	logger  *Logger
	Props   Properties
	DevName string
	isDummy bool
}

// Init called when device first being created before OnStart()
func (dev *Device) Init(name string, logger *Logger, properties []Property) error {
	dev.logger = logger
	dev.Props = NewProperties()
	dev.Props.AddProperties(properties)
	dev.DevName = name
	props := dev.GetProperties()
	dev.isDummy = props.InitProperty("Dummy", "bool", "false", "Is dummy device").(bool)

	dev.LogMessage("Init Device '%s'", dev.Name())
	return nil
}

// GetProperties returns a map of all properties used by device.
// map[string]Property
func (dev *Device) GetProperties() *Properties {
	return &dev.Props
}

// OnStart called once when device first started up. Called after Init()
func (dev *Device) OnStart() error {
	return nil
}

// Name is device name used to identify device
func (dev *Device) Name() string {
	return dev.DevName
}

// Name is device name used to identify device
func (dev *Device) String() string {
	return dev.Name()
}

// OnStop calle once when device is being turned off or deactivated
func (dev *Device) OnStop() error {
	return nil
}

// IsDummyDevice return true if dummy device. False otherwise
func (dev *Device) IsDummyDevice() bool {
	return dev.isDummy
}

// SendNotification Sends messages to devices. Device can monitor notifications
// and handle when needed
func (dev *Device) SendNotification(notify string) error {
	return nil
}

// LogMessage is convenience wrapper for logger
func (dev *Device) LogMessage(pattern string, args ...interface{}) error {
	dev.logger.LogMessage(pattern, args...)
	return nil
}

// LogWarning is convenience wrapper for logger
func (dev *Device) LogWarning(pattern string, args ...interface{}) error {
	dev.logger.LogWarning(pattern, args...)
	return nil
}

// LogError is convenience wrapper for logger
func (dev *Device) LogError(pattern string, args ...interface{}) error {
	dev.logger.LogError(pattern, args...)
	return nil
}

// LogDebug is convenience wrapper for logger
func (dev *Device) LogDebug(pattern string, args ...interface{}) error {
	dev.logger.LogDebug(pattern, args...)
	return nil
}
