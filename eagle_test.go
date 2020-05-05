package engle

import (
	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v2"
	"syscall"
	"testing"
)

func TestConfigFile_Read(t *testing.T) {
	convey.Convey("TestConfigFile_Read", t, func() {
		cF := &configFile{
			FileName: "./testdata/configfile.json",
			FileType: JSON,
		}
		read, _ := cF.Read()
		convey.So(string(read) == "[]", convey.ShouldBeTrue)
	})
}

func TestFileExist(t *testing.T) {
	convey.Convey("TestFileExist", t, func() {
		convey.So(fileExist("./testdata/configfile.json"), convey.ShouldBeTrue)
	})
}

func TestNew(t *testing.T) {
	convey.Convey("TestNew opt set", t, func() {
		engine, _ := New(func(option *Option) error {
			option.Env = "test"
			option.BasePath = "./testdata"
			return nil
		})
		convey.So(engine.opt.Env == "test", convey.ShouldBeTrue)
	})

	convey.Convey("TestNew env set", t, func() {
		_ = syscall.Setenv("eagle.env", "local")
		engine, _ := New(func(option *Option) error {
			option.BasePath = "./testdata"
			return nil
		})
		convey.So(engine.opt.Env == "local", convey.ShouldBeTrue)
	})

	convey.Convey("TestNew env file set", t, func() {
		engine, _ := New(func(option *Option) error {
			option.FileName = "application.yaml"
			option.FileType = YAML
			option.BasePath = "./testdata"
			return nil
		})
		convey.So(engine.opt.FileName == "application.yaml" && engine.opt.FileType == YAML, convey.ShouldBeTrue)
	})

	convey.Convey("TestNew env config", t, func() {
		engine, _ := New(func(option *Option) error {
			option.FileName = "application.yaml"
			option.FileType = YAML
			option.Env = "local"
			option.BasePath = "./testdata"
			return nil
		})
		convey.So(len(engine.configFiles) == 2, convey.ShouldBeTrue)
	})
}

func TestYamlParser_Parse(t *testing.T) {
	convey.Convey("TestYamlParser_Parse", t, func() {
		engine, _ := New(func(option *Option) error {
			option.FileName = "application.yaml"
			option.FileType = YAML
			option.Env = "local"
			option.BasePath = "./testdata"
			return nil
		})
		parser := NewYamlParser(engine)
		err := parser.Parse()
		convey.So(err == nil && len(engine.entities) == 2, convey.ShouldBeTrue)
	})
	convey.Convey("TestJsonParser_Parse", t, func() {
		engine, _ := New(func(option *Option) error {
			option.FileName = "application.json"
			option.FileType = JSON
			option.Env = "local"
			option.BasePath = "./testdata"
			return nil
		})
		parser := NewJsonParser(engine)
		err := parser.Parse()
		convey.So(err == nil && len(engine.entities) == 2, convey.ShouldBeTrue)
	})
}

func TestIniParser_Parse(t *testing.T) {
	convey.Convey("TestIniParser_Parse", t, func() {
		engine, _ := New(func(option *Option) error {
			option.FileName = "full.ini"
			option.FileType = INI
			option.Env = "local"
			option.BasePath = "./testdata"
			return nil
		})
		parser := NewIniParser(engine)
		err := parser.Parse()
		convey.So(err == nil && len(engine.entities) == 2, convey.ShouldBeTrue)
	})

	convey.Convey("TestJsonParser_Parse", t, func() {
		engine, _ := New(func(option *Option) error {
			option.FileName = "application.json"
			option.FileType = JSON
			option.Env = "local"
			option.BasePath = "./testdata"
			return nil
		})
		parser := NewJsonParser(engine)
		err := parser.Parse()
		convey.So(err == nil && len(engine.entities) == 2, convey.ShouldBeTrue)
	})
}

func TestEngine_MapAll(t *testing.T) {
	convey.Convey("TestEngine_MapFull", t, func() {
		type TestStruct struct {
			A int `eagle:"a"`
		}
		engine := &Engine{
			entities: make([]map[interface{}]interface{}, 1),
		}
		buffer := make(map[interface{}]interface{})
		_ = yaml.Unmarshal([]byte("A: 1"), &buffer)
		engine.entities[0] = buffer
		testStruct := &TestStruct{}
		err := engine.MapAll(testStruct)
		convey.So(err == nil && testStruct.A == 1, convey.ShouldBeTrue)
	})
}
