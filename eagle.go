package engle

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

type Option struct {
	FileName       string
	Env            string
	FileType       int
	MergeAnonymous bool
	BasePath       string
}

type Engine struct {
	opt         *Option
	configFiles []configFile
	parser      Parser
	entities    []map[interface{}]interface{}
}

type configFile struct {
	FileName string
	FileType int
}

func (c *configFile) Read() ([]byte, error) {
	file, err := ioutil.ReadFile(c.FileName)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

var parserBuilder = map[int]func(e *Engine) Parser{
	YAML: func(e *Engine) Parser {
		return NewYamlParser(e)
	},
	JSON: func(e *Engine) Parser {
		return NewJsonParser(e)
	},
	INI: func(e *Engine) Parser {
		return NewIniParser(e)
	},
}

var configTypeMap = map[int]string{
	YAML: ".yaml",
	JSON: ".json",
	INI:  ".ini",
}

const DefaultFileName = "application"

func New(optFunctions ...func(*Option) error) (*Engine, error) {
	opt := &Option{}
	for _, optFunction := range optFunctions {
		if err := optFunction(opt); err != nil {
			return nil, err
		}
	}

	if opt.BasePath != "" && opt.BasePath[len(opt.BasePath)-1] != '/' {
		opt.BasePath = opt.BasePath + "/"
	}

	env, ok := os.LookupEnv("eagle.env")
	if ok {
		opt.Env = env
	}

	for fileType, suffix := range configTypeMap {
		if opt.FileName == "" {
			fileName := DefaultFileName + suffix
			if !fileExist(DefaultFileName + suffix) {
				continue
			}
			opt.FileName = fileName
			opt.FileType = fileType
			break
		}

		if !strings.HasSuffix(opt.FileName, suffix) {
			continue
		}
		opt.FileType = fileType
		break
	}

	if opt.FileName == "" || !fileExist(opt.BasePath+opt.FileName) {
		return nil, fmt.Errorf("config file not found")
	}

	engine := &Engine{
		opt:         opt,
		configFiles: make([]configFile, 0),
		entities:    make([]map[interface{}]interface{}, 0),
	}

	engine.initEnvConfig()
	engine.addConfigFile(configFile{
		FileName: opt.BasePath + opt.FileName,
		FileType: opt.FileType,
	})
	engine.parser = parserBuilder[opt.FileType](engine)
	if err := engine.Parse(); err != nil {
		return nil, err
	}

	return engine, nil
}

func (e *Engine) initEnvConfig() {
	if e.opt.Env == "" {
		return
	}

	suffix := configTypeMap[e.opt.FileType]
	fileBaseName := e.opt.FileName[:strings.LastIndex(e.opt.FileName, ".")]
	fileName := fmt.Sprintf("%s-%s%s", fileBaseName, e.opt.Env, suffix)
	if !fileExist(e.opt.BasePath + fileName) {
		return
	}

	e.addConfigFile(configFile{
		FileName: fileName,
		FileType: e.opt.FileType,
	})
}

func (e *Engine) addConfigFile(file configFile) {
	e.configFiles = append([]configFile{file}, e.configFiles...)
}

func (e *Engine) Parse() error {
	return e.parser.Parse()
}

func (e *Engine) MapAll(dst interface{}) error {
	typeOf := reflect.TypeOf(dst)
	if typeOf.Kind() != reflect.Ptr {
		return fmt.Errorf("param must pointer")
	}
	typeOf = typeOf.Elem()
	value := reflect.ValueOf(dst)
	value = value.Elem()

	for _, entity := range e.entities {
		err := e.doSolve(typeOf, value, entity, ".", nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) doSolve(dstType reflect.Type, dstValue reflect.Value, curConfig interface{}, path string, dstStructField *reflect.StructField) error {
	if curConfig == nil {
		return nil
	}

	if !dstValue.CanSet() {
		return nil
	}

	curPath := path[1:]
	switch dstValue.Kind() {
	case reflect.Ptr:
		value := reflect.New(dstType.Elem())
		err := e.doSolve(dstType, value.Elem(), curConfig, path, dstStructField)
		if err != nil {
			return err
		}

		dstValue.Set(value)
		return nil

	case reflect.Slice, reflect.Array:
		curValue, ok := curConfig.([]interface{})
		if !ok {
			return NewConvertError(curPath, dstValue.Kind().String())
		}

		l := len(curValue)
		slice := reflect.MakeSlice(dstValue.Type(), l, l)
		for i := 0; i < l; i++ {
			curDstValue := reflect.New(dstValue.Type().Elem()).Elem()
			curDstType := dstValue.Type()
			if dstStructField != nil {
				curPath = curPath + "." + dstStructField.Name
			}

			err := e.doSolve(curDstType, curDstValue, curValue[i], curPath+".", dstStructField)
			slice.Index(i).Set(curDstValue)
			if err != nil {
				return err
			}
		}
		dstValue.Set(slice)

	case reflect.Struct:
		l := dstValue.NumField()
		curMap, ok := curConfig.(map[interface{}]interface{})
		if !ok {
			return NewConvertError(curPath, dstValue.Kind().String())
		}

		for i := 0; i < l; i++ {
			curDstStructField := dstType.Field(i)
			curDstType := curDstStructField.Type
			curDstField := dstValue.Field(i)

			name := dstStructField.Tag.Get("eagle")
			if name == "" {
				name = curDstStructField.Name
			} else if name == "-" {
				continue
			}

			var err error
			if curDstStructField.Anonymous && e.opt.MergeAnonymous {
				err = e.doSolve(curDstType, curDstField, curConfig, curPath+".", &curDstStructField)
			} else {
				err = e.doSolve(curDstType, curDstField, curMap[name], curPath+"."+curDstStructField.Name, &curDstStructField)
			}
			if err != nil {
				return err
			}
		}
		return nil

	case reflect.Map:
		curMap, ok := curConfig.(map[interface{}]interface{})
		if !ok {
			return NewConvertError(curPath, dstValue.Kind().String())
		}

		makeMap := reflect.MakeMap(dstValue.Type())
		for curMapKey, curMapEntity := range curMap {
			makeMap.SetMapIndex(reflect.ValueOf(curMapKey), reflect.ValueOf(curMapEntity))
		}
		dstValue.Set(makeMap)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		curValue, ok := curConfig.(int)
		if !ok {
			return NewConvertError(curPath, dstValue.Kind().String())
		}

		dstValue.SetInt(int64(curValue))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		curValue, ok := curConfig.(int)
		if !ok {
			return NewConvertError(curPath, dstValue.Kind().String())
		}

		dstValue.SetUint(uint64(curValue))

	case reflect.Float32, reflect.Float64:
		curValue, ok := curConfig.(float64)
		if !ok {
			return NewConvertError(curPath, dstValue.Kind().String())
		}

		dstValue.SetFloat(curValue)

	case reflect.String:
		curValue, ok := curConfig.(string)
		if !ok {
			return NewConvertError(curPath, dstValue.Kind().String())
		}

		dstValue.SetString(curValue)
	case reflect.Bool:
		curValue, ok := curConfig.(bool)
		if !ok {
			return NewConvertError(curPath, dstValue.Kind().String())
		}

		dstValue.SetBool(curValue)
	default:
		return fmt.Errorf("%s type not support", curPath)
	}
	return nil
}
