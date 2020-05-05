package engle

import (
	"github.com/yohox/json-ex/json"
)

type JsonParser struct {
	engine     *Engine
	jsonParser Parser
}

func NewJsonParser(e *Engine) *JsonParser {
	return &JsonParser{
		engine: e,
	}
}

func (y JsonParser) Parse() error {
	y.engine.entities = make([]map[interface{}]interface{}, len(y.engine.configFiles))
	for i, configFile := range y.engine.configFiles {
		bytes, err := configFile.Read()
		if err != nil {
			return err
		}
		var entity map[interface{}]interface{}
		if err := json.Unmarshal(bytes, &entity); err != nil {
			return err
		}
		y.engine.entities[i] = entity
	}

	return nil
}
