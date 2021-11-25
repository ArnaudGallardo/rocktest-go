package scenario

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

func (module *Module) Executable(params map[string]interface{}, scenario *Scenario) error {

	paramsEx, err := scenario.ExpandMap(params)
	if err != nil {
		return err
	}

	val, err := scenario.GetString(paramsEx, "value", nil)
	if err != nil {
		return err
	}

	args := strings.Fields(val)
	if len(args) < 1 {
		return errors.New("Executable path should be defined")
	}

	jsonOutput, err := scenario.GetBool(paramsEx, "jsonOutput", false)
	if err != nil {
		return err
	}

	out, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		return err
	}

	if jsonOutput {
		var result map[string]interface{}
		json.Unmarshal([]byte(out), &result)

		for key, value := range result {
			scenario.PutContext(key, value.(string))
		}
	}

	return nil
}
