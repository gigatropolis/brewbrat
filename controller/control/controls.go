package control

import (
	"../config"
)

type IDevice interface {
	Init(name string, logger *Logger, properties []Property) error
	OnStart() error
	OnStop() error
	Name() string
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
}

func (dev *Device) Init(name string, logger *Logger, properties []Property) error {
	dev.logger = logger
	dev.Props = NewProperties()
	dev.Props.AddProperties(properties)
	dev.DevName = name
	dev.LogMessage("Init Device '%s'", dev.Name())
	return nil
}

func (dev *Device) GetProperties() *Properties {
	return &dev.Props
}

func (dev *Device) OnStart() error {
	return nil
}
func (dev *Device) Name() string {
	return dev.DevName
}

func (dev *Device) OnStop() error {
	return nil
}

// LogMessage is wrapper for logger
func (dev *Device) LogMessage(pattern string, args ...interface{}) error {
	dev.logger.LogMessage(pattern, args...)
	return nil
}
func (dev *Device) LogWarning(pattern string, args ...interface{}) error {
	dev.logger.LogWarning(pattern, args...)
	return nil
}
func (dev *Device) LogError(pattern string, args ...interface{}) error {
	dev.logger.LogError(pattern, args...)
	return nil
}
func (dev *Device) LogDebug(pattern string, args ...interface{}) error {
	dev.logger.LogDebug(pattern, args...)
	return nil
}
