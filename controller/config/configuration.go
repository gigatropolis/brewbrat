package config

import (
	"encoding/xml"
	//"../control"
)

// BrewController is all configured devices
// Represents the configuration file
type BrewController struct {
	XMLName    xml.Name          `xml:"controller"`
	Name       string            `xml:"name"`
	Version    string            `xml:"version"`
	Equipment  []EquipmentConfig `xml:"equipment>equip"`
	Sensors    []SensorConfig    `xml:"sensors>sensor"`
	Actors     []ActorsConfig    `xml:"actors>actor"`
	Properties []PropertyConfig  `xml:"properties>property"`
}

// EquipmentConfig is a kettle, mashtun, etc.
// reads values from sensors and sets the actors
type EquipmentConfig struct {
	XMLName    xml.Name         `xml:"sensor"`
	Name       string           `xml:"name"`
	Type       string           `xml:"type"`
	Properties []PropertyConfig `xml:"properties>property"`
}

// SensorConfig a sensor that reads a value from divice
type SensorConfig struct {
	XMLName    xml.Name         `xml:"sensor"`
	Name       string           `xml:"name"`
	Type       string           `xml:"type"`
	Properties []PropertyConfig `xml:"properties>property"`
}

// ActorsConfig is type of relay or any on/off device
type ActorsConfig struct {
	XMLName    xml.Name         `xml:"actor"`
	Name       string           `xml:"name"`
	Type       string           `xml:"type"`
	Properties []PropertyConfig `xml:"properties>property"`
}

// PropertyConfig are the attribute values for devices
// passed in by the configuration
type PropertyConfig struct {
	XMLName xml.Name `xml:"property"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
	Hidden  bool     `xml:"hidden,attr"`
	Comment string   `xml:"comment,attr"`
	Choice  string   `xml:"choice,attr"`
	Select  string   `xml:"select,attr"`
	Value   string   `xml:",chardata"`
}

// DefaultSensorConfig is temp code to return initial sensor devices for demo
func DefaultSensorConfig() ([]SensorConfig, error) {
	sensorsDefined := []SensorConfig{
		{Name: "temp Sensor 1",
			Type: "TempSensor",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: "temp Sensor 1", Comment: "Sensor Name", Choice: ""},
				{Name: "Address", Type: "uint", Hidden: false, Value: "7205759448148251176", Comment: "1-Wire sensor address", Choice: ""},
				{Name: "Units", Type: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""}},
		},
		{Name: "Dummy Temp 1",
			Type: "DummyTempSensor",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: "Dummy Temp 1", Comment: "Sensor Name", Choice: ""},
				{Name: "Units", Type: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""}},
		},
	}
	return sensorsDefined, nil
}

// DefaultRelayConfig is temp code to return initial relay devices for demo
func DefaultRelayConfig() ([]ActorsConfig, error) {
	relayDefined := []ActorsConfig{
		{Name: "Relay 1",
			Type: "SimpleRelay",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: "Relay 1", Comment: "relay Name", Choice: ""},
				{Name: "GPIO", Type: "string", Hidden: false, Value: "P1_38", Comment: "GPIO by name", Choice: ""},
			},
		},
		{Name: "Relay 2",
			Type: "SimpleRelay",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: "Relay 2", Comment: "relay Name", Choice: ""},
				{Name: "GPIO", Type: "string", Hidden: false, Value: "P1_40", Comment: "GPIO by name", Choice: ""},
			},
		},
		{Name: "SSR 1",
			Type: "SimpleSSR",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: "SSR 1", Comment: "SSR Name", Choice: ""},
				{Name: "GPIO", Type: "string", Hidden: false, Value: "P1_36", Comment: "GPIO by name", Choice: ""},
			},
		},
	}
	return relayDefined, nil
}
