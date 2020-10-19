package config

import (
	"encoding/xml"
	"fmt"
	"strconv"
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

func DefaultEquipment(dummy bool) ([]EquipmentConfig, error) {
	eq := []EquipmentConfig{}

	if dummy {
		eq = append(eq, EquipmentConfig{
			Name: "Dummy Equipment 1",
			Type: "SimpleRIMM",
			Properties: []PropertyConfig{
				{Name: "Temp Sensor", Type: "string", Hidden: false, Value: "Temp Sensor 1", Comment: "Sensor Name", Choice: ""},
				{Name: "Units", Type: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""},
				{Name: "Pump", Type: "string", Hidden: false, Value: "Dummy Relay 1", Comment: "Units for Sensor", Choice: ""},
				{Name: "Circulator", Type: "string", Hidden: false, Value: "Dummy Relay 2", Comment: "Units for Sensor", Choice: ""},
				{Name: "Heater", Type: "string", Hidden: false, Value: "Dummy Relay 3", Comment: "Units for Sensor", Choice: ""}},
		})

	} else {
		eq = append(eq, EquipmentConfig{
			Name: "Equipment 1",
			Type: "SimpleRIMM",
			Properties: []PropertyConfig{
				{Name: "Temp Sensor", Type: "string", Hidden: false, Value: "Dummy Temp 1", Comment: "Sensor Name", Choice: ""},
				{Name: "Units", Type: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""},
				{Name: "Pump", Type: "string", Hidden: false, Value: "Relay 1", Comment: "Units for Sensor", Choice: ""},
				{Name: "Circulator", Type: "string", Hidden: false, Value: "Relay 2", Comment: "Units for Sensor", Choice: ""},
				{Name: "Heater", Type: "string", Hidden: false, Value: "SSR 1", Comment: "Units for Sensor", Choice: ""}},
		})
	}
	return eq, nil
}

// DefaultSensorConfig is temp code to return initial sensor devices for demo
func DefaultSensorConfig(adrs []uint64, dummy bool) ([]SensorConfig, error) {
	tempNum := uint64(1)
	sensorsDefined := []SensorConfig{}

	if dummy {
		for i := 1; i <= 3; i++ {
			sNum := strconv.FormatInt(1, 10)
			sensorsDefined = append(sensorsDefined, SensorConfig{
				Name: "Dummy Temp " + sNum,
				Type: "DummyTempSensor",
				Properties: []PropertyConfig{
					{Name: "Name", Type: "string", Hidden: false, Value: "Dummy Temp " + sNum, Comment: "Sensor Name", Choice: ""},
					{Name: "Units", Type: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""}},
			})
		}
	} else {
		for _, adr := range adrs {
			sTempName := "temp Sensor " + strconv.FormatUint(tempNum, 10)
			sTempAdress := strconv.FormatUint(adr, 10)
			tempNum++
			fmt.Printf("config Address = %s\n", sTempAdress)

			sensorsDefined = append(sensorsDefined, SensorConfig{
				Name: sTempName,
				Type: "TempSensor",
				Properties: []PropertyConfig{
					{Name: "Name", Type: "string", Hidden: false, Value: sTempName, Comment: "Sensor Name", Choice: ""},
					{Name: "Address", Type: "uint", Hidden: false, Value: sTempAdress, Comment: "1-Wire sensor address", Choice: ""},
					{Name: "Units", Type: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""}},
			})
		}

		sensorsDefined = append(sensorsDefined, SensorConfig{
			Name: "Dummy Temp 1",
			Type: "DummyTempSensor",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: "Dummy Temp 1", Comment: "Sensor Name", Choice: ""},
				{Name: "Units", Type: "string", Hidden: false, Value: "°F", Comment: "Units for Sensor", Choice: ""}},
		})
	}

	return sensorsDefined, nil
}

func DefaultConfiguration(adrs []uint64, dummy bool) (BrewController, error) {
	brewController := BrewController{}
	var err error
	brewController.Sensors, err = DefaultSensorConfig(adrs, dummy)
	brewController.Actors, err = DefaultRelayConfig()
	brewController.Equipment, err = DefaultEquipment(dummy)
	return brewController, err
}

// DefaultRelayConfig is temp code to return initial relay devices for demo
func DefaultRelayConfig() ([]ActorsConfig, error) {
	relayDefined := []ActorsConfig{
		{Name: "Relay 1",
			Type: "SimpleRelay",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: "Relay 1", Comment: "relay Name", Choice: ""},
				{Name: "GPIO", Type: "string", Hidden: false, Value: "GPIO21", Comment: "GPIO by name", Choice: ""},
			},
		},
		{Name: "Relay 2",
			Type: "SimpleRelay",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: "Relay 2", Comment: "relay Name", Choice: ""},
				{Name: "GPIO", Type: "string", Hidden: false, Value: "GPIO20", Comment: "GPIO by name", Choice: ""},
			},
		},
		{Name: "SSR 1",
			Type: "SimpleSSR",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: "SSR 1", Comment: "SSR Name", Choice: ""},
				{Name: "GPIO", Type: "string", Hidden: false, Value: "GPIO16", Comment: "GPIO by name", Choice: ""},
			},
		},
		{Name: "Dummy Relay 1",
			Type: "DummyRelay",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: "Dummy Relay 1", Comment: "Dummy relay Name", Choice: ""},
			},
		},
		{Name: "Dummy Relay 2",
			Type: "DummyRelay",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: "Dummy Relay 2", Comment: "Dummy relay Name", Choice: ""},
			},
		},
		{Name: "Dummy Relay 3",
			Type: "DummyRelay",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: "Dummy Relay 3", Comment: "Dummy relay Name", Choice: ""},
			},
		},
	}
	return relayDefined, nil
}
