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
	Buzzers    []BuzzerConfig    `xml:"buzzers>buzzer"`
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

// BuzzerConfig is a buzzer device
type BuzzerConfig struct {
	XMLName    xml.Name         `xml:"buzzer"`
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
			sNum := strconv.FormatInt(int64(i), 10)
			sensorsDefined = append(sensorsDefined, SensorConfig{
				Name: "temp Sensor " + sNum,
				Type: "DummyTempSensor",
				Properties: []PropertyConfig{
					{Name: "Name", Type: "string", Hidden: false, Value: "temp Sensor " + sNum, Comment: "Sensor Name", Choice: ""},
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

// DefaultConfiguration Creates a default configuration object
func DefaultConfiguration(adrs []uint64, relayGPIO []string, ssrGPIO []string, dummy bool) (BrewController, error) {
	brewController := BrewController{}
	var err error
	brewController.Buzzers, err = DefaultBuzzerConfig(dummy)
	brewController.Sensors, err = DefaultSensorConfig(adrs, dummy)
	brewController.Actors, err = DefaultRelayConfig(relayGPIO, ssrGPIO, dummy)
	brewController.Equipment, err = DefaultEquipment(dummy)
	return brewController, err
}

// DefaultRelayConfig is temp code to return initial relay devices for demo
func DefaultRelayConfig(relayGPIO []string, ssrGPIO []string, dummy bool) ([]ActorsConfig, error) {
	relayDefined := []ActorsConfig{}

	relayType := "SimpleRelay"
	ssType := "SimpleSSR"
	if dummy {
		relayType = "DummyRelay"
		ssType = "DummyRelay"
	}

	for indx, rel := range relayGPIO {
		name := fmt.Sprintf("Relay %d", indx+1)
		relayDefined = append(relayDefined, ActorsConfig{Name: name,
			Type: relayType,
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: name, Comment: "relay Name", Choice: ""},
				{Name: "GPIO", Type: "string", Hidden: false, Value: rel, Comment: "GPIO by name", Choice: ""},
			},
		})
	}
	for indx, ssrs := range ssrGPIO {
		name := fmt.Sprintf("SSR %d", indx+1)
		relayDefined = append(relayDefined, ActorsConfig{
			Name: name,
			Type: ssType,
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: name, Comment: "SSR Name", Choice: ""},
				{Name: "GPIO", Type: "string", Hidden: false, Value: ssrs, Comment: "GPIO by name", Choice: ""},
			},
		})
	}
	for indx := 1; indx <= 4; indx++ {
		name := fmt.Sprintf("Dummy Relay %d", indx)
		relayDefined = append(relayDefined, ActorsConfig{
			Name: name,
			Type: "DummyRelay",
			Properties: []PropertyConfig{
				{Name: "Name", Type: "string", Hidden: false, Value: name, Comment: "Dummy relay Name", Choice: ""},
			},
		})
	}
	return relayDefined, nil
}

// DefaultBuzzerConfig is temp code to return initial buzzer device for demo
func DefaultBuzzerConfig(dummy bool) ([]BuzzerConfig, error) {
	buzzerDefined := []BuzzerConfig{}

	buzzerType := "ActiveBuzzer"
	if dummy {
		buzzerType = "DummyBuzzer"
	}

	name := "Main Buzzer"
	buzzerDefined = append(buzzerDefined, BuzzerConfig{
		Name: name,
		Type: buzzerType,
		Properties: []PropertyConfig{
			{Name: "Name", Type: "string", Hidden: false, Value: name, Comment: "Buzzer Name", Choice: ""},
		},
	})

	return buzzerDefined, nil
}
