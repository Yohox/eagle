package engle

import "gopkg.in/yaml.v2"

type YamlParser struct {
	engine     *Engine
	jsonParser Parser
}

func NewYamlParser(e *Engine) *YamlParser {
	return &YamlParser{
		engine: e,
	}
}

func (y YamlParser) Parse() error {
	y.engine.entities = make([]map[interface{}]interface{}, len(y.engine.configFiles))
	for i, configFile := range y.engine.configFiles {
		bytes, err := configFile.Read()
		if err != nil {
			return err
		}
		var entity map[interface{}]interface{}
		if err := yaml.Unmarshal(bytes, &entity); err != nil {
			return err
		}
		y.engine.entities[i] = entity
	}

	return nil
}
