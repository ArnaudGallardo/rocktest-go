package scenario

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
	"io.rocktest/rocktest/text"

	log "github.com/sirupsen/logrus"
)

type Scenario struct {
	Steps []interface{}

	// Contains the variables
	Context map[string]string
	Subst   *text.StringSubstitutor

	// Caller, when the scenario is called from another one
	Caller *Scenario

	// Root, from where looking from modules
	Root string

	Skip bool
}

func NewScenario() *Scenario {
	ret := new(Scenario)
	ret.Context = make(map[string]string)
	ret.Subst = text.NewStringSubstitutorByMap(ret.Context)
	ret.Skip = false
	return ret
}

func nodeToList(node interface{}) []interface{} {
	m, ok := node.([]interface{})
	if !ok {
		panic(fmt.Sprintf("%v is not of type list", node))
	}
	return m
}

func nodeToMap(node interface{}) map[string]interface{} {
	m, ok := node.(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("%v is not of type map", node))
	}
	return m
}

func humanStr(src interface{}) string {
	if src != nil {
		return fmt.Sprint(src)
	} else {
		return ""
	}
}

func (s *Scenario) isBuiltin(name string) bool {
	return name == "module" || name == "step"
}

func (s *Scenario) RunSteps(steps []interface{}) error {
	var i int = 0
	var stop bool = false

	for _, v := range steps {
		i++

		s.Context["step"] = fmt.Sprint(i)

		node := nodeToMap(v)
		step := NewStep(node, s)

		if !s.Skip && !strings.HasPrefix(step.Type, "--") {
			log.Infof("[%s/%d] %s - %s", s.Context["module"], i, strings.ToUpper(step.Type), humanStr(step.Params["value"]))
			if log.IsLevelEnabled(log.DebugLevel) {
				for name, val := range step.Params {
					log.Debugf("  %s = %s", name, val)
				}
			}
		}

		switch step.Type {
		case "Exit":
			stop = true
		default:

			if step.Type == "Resume" || (!s.Skip && !strings.HasPrefix(step.Type, "--")) {
				err := step.Exec()

				if err != nil {

					yamlMap := make(map[string]interface{})
					yamlMap["stepNumber"] = i
					yamlMap["scenario"] = s.Context["module"]
					if step.Desc != "" {
						yamlMap["desc"] = step.Desc
					}
					yamlMap["error"] = err.Error()
					yamlMap["step"] = node

					yamlString, _ := yaml.Marshal(yamlMap)

					return errors.New(string(yamlString))
				}

			}
		}

		if stop {
			break
		}
	}

	return nil
}

func (s *Scenario) RunFromRoot(scen string) error {
	log.Infof("Run scenario %s", scen)

	yamlFile, err := ioutil.ReadFile(s.Root + "/" + scen)
	if err != nil {
		return err
	}

	basename := filepath.Base(scen)

	s.Context["module"] = strings.TrimSuffix(basename, filepath.Ext(basename))

	err = yaml.Unmarshal(yamlFile, &s.Steps)

	if err != nil {
		return err
	}

	steps := nodeToList(s.Steps)

	return s.RunSteps(steps)

}

func (s *Scenario) Run(scen string) error {
	log.Infof("Run scenario %s", scen)

	yamlFile, err := ioutil.ReadFile(scen)
	if err != nil {
		return err
	}

	basename := filepath.Base(scen)

	abs, _ := filepath.Abs(scen)

	s.Root = filepath.Dir(abs)
	s.Context["module"] = strings.TrimSuffix(basename, filepath.Ext(basename))

	err = yaml.Unmarshal(yamlFile, &s.Steps)

	if err != nil {
		return err
	}

	steps := nodeToList(s.Steps)

	return s.RunSteps(steps)

}

func (s *Scenario) PutContext(name string, value interface{}) error {

	log.Debugf("Set %s: %s = %v", s.Context["module"], name, value)

	switch str := value.(type) {
	case string:
		s.Context[name] = str
	case int:
		s.Context[name] = fmt.Sprint(str)
	default:
		return fmt.Errorf("variable value type must be string or int, not %T", value)
	}

	return nil
}

func (s *Scenario) AddVariables(params map[string]interface{}) error {

	for k, v := range params {

		err := s.PutContext(k, v)
		if err != nil {
			return err
		}

	}

	return nil
}

func (s *Scenario) CopyVariable(name string, source *Scenario) error {

	val, found := source.Context[name]
	if found {
		err := s.PutContext(name, val)
		if err != nil {
			return err
		}
	}

	return nil
}

// Add all the variables from params, excluding builtin variables
func (s *Scenario) CopyVariables(source *Scenario) error {

	for k, v := range source.Context {

		if s.isBuiltin(k) {
			continue
		}

		err := s.PutContext(k, v)
		if err != nil {
			return err
		}

	}

	return nil
}

// Replace the values of the variables in a map
func (s *Scenario) ExpandMap(params map[string]interface{}) map[string]interface{} {

	ret := make(map[string]interface{})

	for k, v := range params {

		switch v.(type) {
		case string:
			ret[k] = s.Subst.Replace(params[k].(string))
		default:
			ret[k] = params[k]
		}

	}

	return ret

}

// Get a parameter as string. Returns the default value if not found.
// If the value is not found, and there is no default value, returns an error.
// If the value is not a string, return an error
func (s *Scenario) GetString(params map[string]interface{}, key string, def interface{}) (string, error) {

	if params == nil {
		if def == nil {
			return "", errors.New("Params map empty, and no default value provided for key " + key)
		} else {
			return def.(string), nil
		}
	}

	ret, ok := params[key]

	if ok {
		switch ret := ret.(type) {
		case string:
			return ret, nil
		case int:
			return fmt.Sprint(ret), nil
		default:
			msg := fmt.Sprintf("Bad type for value %s. Must be string, not %v", key, reflect.TypeOf(ret))
			return "", errors.New(msg)
		}

	} else {
		if def == nil {
			return "", errors.New("Value not found for " + key)
		} else {
			return def.(string), nil
		}
	}

}

// Get a parameter as list. Returns the default value if not found.
// If the value is not found, and there is no default value, returns an error.
// If the value is not a list, return an error
func (s *Scenario) GetList(params map[string]interface{}, key string, def interface{}) ([]interface{}, error) {

	if params == nil {
		if def == nil {
			return nil, errors.New("Params map empty, and no default value provided for key " + key)
		} else {
			return def.([]interface{}), nil // List of the builtin variable
		}
	}

	ret, ok := params[key]

	if ok {
		switch ret := ret.(type) {
		case []interface{}:
			return ret, nil
		default:
			msg := fmt.Sprintf("Bad type for value %s. Must be a list, not %v", key, reflect.TypeOf(ret))
			return nil, errors.New(msg)
		}

	} else {
		if def == nil {
			return nil, errors.New("Value not found for " + key)
		} else {
			switch ret := def.(type) {
			case []interface{}:
				return ret, nil
			default:
				msg := fmt.Sprintf("Bad type for default value for %s. Must be a list, not %v", key, reflect.TypeOf(ret))
				return nil, errors.New(msg)
			}
		}
	}

}
