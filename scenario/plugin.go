package scenario

import (
	"errors"
	"plugin"
	"fmt"
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

	fmt.Println("toto")
	p, err := plugin.Open(val)
	if err != nil {
		panic(err)
	}
	fmt.Println("test")
	v, err := p.Lookup("Name")
	if err != nil {
		return errors.New("Plugin is missing the Name variable.")
	}
	fmt.Println(*v.(*string))
	/*
	f, err := p.Lookup(*v.(*string))
	if err != nil {
		panic(err)
	}
	f.(func(map[string]interface {}) error)(params) // prints "Hello, number 7"

	*/

	return nil
}
