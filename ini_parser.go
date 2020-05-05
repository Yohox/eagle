package engle

import (
	"gopkg.in/ini.v1"
)

type IniParser struct {
	engine    *Engine
	iniParser Parser
}

func NewIniParser(e *Engine) *IniParser {
	return &IniParser{
		engine: e,
	}
}

func (i IniParser) Parse() error {
	i.engine.entities = make([]map[interface{}]interface{}, len(i.engine.configFiles))
	for index, configFile := range i.engine.configFiles {
		bytes, err := configFile.Read()
		if err != nil {
			return err
		}
		var entity map[interface{}]interface{}
		cfg, err := ini.Load(bytes)
		if err != nil {
			return err
		}
		if err := cfg.MapTo(&entity); err != nil {
			return err
		}
		i.engine.entities[index] = entity
	}

	return nil
}
