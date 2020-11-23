package gocfg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

// TODO: add config size limit

const EnvPrefix = "Env"

type ICfg interface {
	Load(path string, config interface{}) error
	Bool(key string) bool
	Int(key string) int
	Float(key string) float64
	String(key string) string
	Interface(key string) interface{}
	SetBool(key string, val bool)
	SetInt(key string, val int)
	SetFloat(key string, val float64)
	SetString(key string, val string)
	SetInterface(key string, val interface{})
	Print()
}

type Cfg struct {
	BoolVals      map[string]bool
	IntVals       map[string]int
	FloatVals     map[string]float64
	StringVals    map[string]string
	InterfaceVals map[string]interface{}
}

func New() *Cfg {
	return &Cfg{
		BoolVals:      map[string]bool{},
		IntVals:       map[string]int{},
		FloatVals:     map[string]float64{},
		StringVals:    map[string]string{},
		InterfaceVals: map[string]interface{}{},
	}
}

func (c *Cfg) Load(path string, config interface{}) error {
	cfgObj, err := load(path, config)
	if err != nil {
		return err
	}
	return c.visit(cfgObj)
}

func (c *Cfg) Print() {
	for k, v := range c.BoolVals {
		fmt.Printf("\nboolVal: %s=%t", k, v)
	}
	for k, v := range c.IntVals {
		fmt.Printf("\nintlVal: %s=%d", k, v)
	}
	for k, v := range c.FloatVals {
		fmt.Printf("\nfloatVal: %s=%f", k, v)
	}
	for k, v := range c.StringVals {
		fmt.Printf("\nstringVal: %s=%s", k, v)
	}
	for k, v := range c.InterfaceVals {
		fmt.Printf("\ninterfaceVal: %s=%v", k, v)
	}
}

func load(path string, config interface{}) (interface{}, error) {
	cfgFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	cfgBytes, err := ioutil.ReadAll(cfgFile)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(cfgBytes, config); err != nil {
		return nil, err
	}
	return config, nil
}

type ValueInfo struct {
	v    reflect.Value
	path string
	name string
}

func (c *Cfg) visit(cfgObj interface{}) error {
	queue := []*ValueInfo{}
	queue = append(
		queue,
		&ValueInfo{
			v:    reflect.ValueOf(cfgObj).Elem(),
			name: "",
			path: "",
		},
	)

	for len(queue) > 0 {
		e := queue[0]
		queue = queue[1:]

		k := e.v.Kind()
		switch {
		case strings.HasPrefix(e.name, EnvPrefix):
			// must be a string
			envName := e.v.String()
			envValue, ok := os.LookupEnv(envName)
			if !ok {
				return fmt.Errorf("cfg: environment is not set: (%s)=(%s)", e.name, envName)
			}
			c.StringVals[e.path] = envValue
		case k == reflect.Bool:
			c.BoolVals[e.path] = e.v.Bool()
		case k == reflect.Int:
			c.IntVals[e.path] = e.v.Interface().(int) // use int instead of uint/int/8/16/32/64
		case k == reflect.Float64:
			c.FloatVals[e.path] = e.v.Float()
		case k == reflect.String:
			c.StringVals[e.path] = e.v.String()
		case k == reflect.Slice || k == reflect.Map:
			c.InterfaceVals[e.path] = e.v.Interface()
		case k == reflect.Struct || k == reflect.Ptr:
			obj := e.v
			if k == reflect.Ptr {
				obj = e.v.Elem()
			}
			for i := 0; i < obj.NumField(); i++ {
				childName := obj.Type().Field(i).Name
				childValue := obj.Field(i)
				info := &ValueInfo{
					v:    childValue,
					name: childName,
					path: fmt.Sprintf("%s.%s", e.path, childName),
				}
				queue = append(queue, info)
			}
		default:
			return fmt.Errorf("cfg: %s type %s is not supproted", e.path, k)
		}
	}

	return nil
}

func (c *Cfg) Bool(key string) bool             { return c.BoolVals[key] }
func (c *Cfg) Int(key string) int               { return c.IntVals[key] }
func (c *Cfg) Float(key string) float64         { return c.FloatVals[key] }
func (c *Cfg) String(key string) string         { return c.StringVals[key] }
func (c *Cfg) Interface(key string) interface{} { return c.InterfaceVals[key] }

func (c *Cfg) SetBool(key string, val bool)             { c.BoolVals[key] = val }
func (c *Cfg) SetInt(key string, val int)               { c.IntVals[key] = val }
func (c *Cfg) SetFloat(key string, val float64)         { c.FloatVals[key] = val }
func (c *Cfg) SetString(key string, val string)         { c.StringVals[key] = val }
func (c *Cfg) SetInterface(key string, val interface{}) { c.InterfaceVals[key] = val }
