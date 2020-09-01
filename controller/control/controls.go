package control

type IDevice interface {
	Init(logger *Logger, properties []Property) error
	OnStart() error
	OnStop() error
	LogMessage(pattern string, args ...interface{}) error
	LogWarning(pattern string, args ...interface{}) error
	LogError(pattern string, args ...interface{}) error
	LogDebug(pattern string, args ...interface{}) error
}

type IActor interface {
	IDevice
	On() error
	Off() error
	SetPower(int64 power) error
}

type EqMessage struct {
	DeviceName string
	cmd int64
	value float64
}

type IEquipment interface {
	IDevice
	Init(logger *Logger, properties []Property, <-in chan EqMessage, out<- chan EqMessage)
	Run() error
}

type Device struct {
	logger *Logger
	Props  Properties
}

func (dev *Device) Init(logger *Logger, properties []Property) error {
	dev.logger = logger
	dev.Props = NewProperties()
	dev.Props.AddProperties(properties)
	dev.LogMessage("Init Device...")
	return nil
}

func (dev *Device) GetProperties() *Properties {
	return &dev.Props
}

func (dev *Device) OnStart() error {
	return nil
}

func (dev *Device) OnStop() error {
	return nil
}

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

type Actor struct {
	Device
}

func (act *Actor) Init(logger *Logger, properties []Property) error {
	act.Device.Init(logger, properties)
	return nil
}

func (act *Actor) On() error {
	return nil
}

func (act *Actor) Off() error {
	return nil
}

func (act *Actor) SetPower(int64 power) error {
	return nil
}

type BrewController struct {
	XMLName      xml.Name         `xml:"controller"`
	Version      string           `xml:"version"`
	Sensors      []SensorConfig   `xml:"sensors>sensor"`
	Actors       []ActorsConfig   `xml:"actors>actor"`
	Property   []PropertyConfig    `xml:"properties>property`
}

type SensorConfig struct {
	XMLName      xml.Name         `xml:"sensor"`
	Name           string          `xml:"name"`
	Type           string          `xml:"type"`
	Property   []PropertyConfig    `xml:"properties>property`

}

type ActorsConfig struct {
	XMLName      xml.Name         `xml:"actor"`
	Name           string          `xml:"name"`
	Type           string          `xml:"type"`
	Property   []PropertyConfig    `xml:"properties>property`
}

type PropertyConfig struct {
	XMLName      xml.Name         `xml:"property"`
	Name     string          `xml:"name,attr"`
	Type     string          `xml:"type,attr"`
	Hidden   bool            `xml:"hidden,attr"`
	Comment  string          `xml:"comment,attr"`
	Value    string          `xml:",chardata"`
}
