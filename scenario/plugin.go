package scenario

import (
	"errors"
	"plugin"
)

func (module *Module) Plugin(params map[string]interface{}, scenario *Scenario) error {

	paramsEx, err := scenario.ExpandMap(params)
	if err != nil {
		return err
	}

	val, err := scenario.GetString(paramsEx, "value", nil)
	if err != nil {
		return err
	}

	p, err := plugin.Open(val)
	if err != nil {
		panic(err)
	}
	v, err := p.Lookup("Name")
	if err != nil {
		return errors.New("Plugin is missing the Name variable.")
	}
	f, err := p.Lookup(*v.(*string))
	if err != nil {
		panic(err)
	}
	f.(func(map[string]interface{}, *Scenario))(params, scenario) // prints "Hello, number 7"

	return nil
}
