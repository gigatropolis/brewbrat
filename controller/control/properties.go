package control

import (
	"fmt"
)

type Property struct {
	Name     string
	PropType string
	Value    interface{}
	Comment  string
}

func (p *Property) String() string {
	return fmt.Sprintf("%v", p.Value)
}

type Properties map[string]Property

func (p *Properties) Count() int {
	return len(*p)
}

func (p *Properties) updatePropertyValue(name string, value interface{}) interface{} {
	if _, ok := (*p)[name]; ok {
		(*p)[name] = Property{name, (*p)[name].PropType, value, (*p)[name].Comment}
	}
	return (*p)[name].Value
}

func (p *Properties) InitProperty(name string, proptype string, value interface{}, comment string) interface{} {

	if _, present := (*p)[name]; !present {
		p.AddProperty(name, proptype, value, comment)
	} else {
		if (*p)[name].Value == nil && value != nil {
			p.updatePropertyValue(name, value)
		}
	}
	return (*p)[name].Value
}

// AddProperty to the list of Properties. Will overwrite if already exists
// returns the value that is added as interface{}
func (p *Properties) AddProperty(name string, proptype string, value interface{}, comment string) interface{} {
	(*p)[name] = Property{name, proptype, value, comment}
	return (*p)[name].Value
}

// AddProperties to the list of Properties. Will overwrite if already exists
// returns the value that is added as interface{}
func (p *Properties) AddProperties(props []Property) {
	for _, prop := range props {
		p.AddProperty(prop.Name, prop.PropType, prop.Value, prop.Comment)
	}
}

// GetParam will return a Param object based on the param name passed in
// ok is True if the param exists. False if the param doesn't exist
func (p *Properties) GetProperty(name string) (Property, bool) {
	val, ok := (*p)[name]
	return val, ok
}

func (p *Properties) GetPropertyType(name string) (string, bool) {
	param, ok := (*p)[name]
	if !ok {
		return "", false
	}
	return param.PropType, true
}

func (p *Properties) GetPropertyValue(name string) (interface{}, bool) {
	param, ok := (*p)[name]
	if !ok {
		return nil, false
	}
	return param.Value, true
}

func (p *Properties) GetPropertyComment(name string) (interface{}, bool) {
	param, ok := (*p)[name]
	if !ok {
		return nil, false
	}
	return param.Comment, true
}

// the CreateProperties returns new empty Properties object
func NewProperties() Properties {
	return Properties{}
}