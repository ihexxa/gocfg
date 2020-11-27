package gocfg

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

var GocfgTag = "cfg"
var GocfgValEnv = "env"
var GocfgValRequired = "required"

// ICfg is an interface defined for consumer according to *gocfg.Cfg
type ICfg interface {
	Bool(key string) (bool, bool)
	Int(key string) (int, bool)
	Float(key string) (float64, bool)
	String(key string) (string, bool)
	Map(key string) (interface{}, bool)
	Slice(key string) (interface{}, bool)
	Struct(key string) (interface{}, bool)

	BoolOr(key string, defaultVal bool) bool
	IntOr(key string, defaultVal int) int
	FloatOr(key string, defaultVal float64) float64
	StringOr(key string, defaultVal string) string
	MapOr(key string, defaultVal interface{}) interface{}
	SliceOr(key string, defaultVal interface{}) interface{}
	StructOr(key string, defaultVal interface{}) interface{}

	GrabBool(key string) bool
	GrabInt(key string) int
	GrabFloat(key string) float64
	GrabString(key string) string
	GrabMap(key string) interface{}
	GrabSlice(key string) interface{}
	GrabStruct(key string) interface{}

	SetBool(key string, val bool)
	SetInt(key string, val int)
	SetFloat(key string, val float64)
	SetString(key string, val string)
	SetStruct(key string, val interface{})

	Load(pvd CfgProvider, config interface{}) error
	Print()
	Debug()
}

// Cfg is an abstraction over a configuration
type Cfg struct {
	debug bool

	boolVals   map[string]bool
	intVals    map[string]int
	floatVals  map[string]float64
	stringVals map[string]string
	mapVals    map[string]interface{}
	sliceVals  map[string]interface{}
	structVals map[string]interface{}
}

type valueInfo struct {
	v    reflect.Value
	path string
	name string
}

// New returns a new *Cfg
func New() *Cfg {
	return &Cfg{
		debug:      false,
		boolVals:   map[string]bool{},
		intVals:    map[string]int{},
		floatVals:  map[string]float64{},
		stringVals: map[string]string{},
		mapVals:    map[string]interface{}{},
		sliceVals:  map[string]interface{}{},
		structVals: map[string]interface{}{},
	}
}

// Debug opens debug mode and prints more logs.
func (c *Cfg) Debug() {
	c.debug = true
}

// Load loads configuration from local path according to config's definition
func (c *Cfg) Load(pvd CfgProvider, config interface{}) error {
	err := pvd.Load(config)
	if err != nil {
		return err
	}
	return c.visit(config)
}

func (c *Cfg) warnf(format string, vals ...interface{}) {
	if c.debug {
		fmt.Printf(format, vals...)
	}
}

// Print prints all of the values in the Cfg
func (c *Cfg) Print() {
	for k, v := range c.boolVals {
		fmt.Printf("\n%s:bool = %t", k, v)
	}
	for k, v := range c.intVals {
		fmt.Printf("\n%s:int = %d", k, v)
	}
	for k, v := range c.floatVals {
		fmt.Printf("\n%s:float = %f", k, v)
	}
	for k, v := range c.stringVals {
		fmt.Printf("\n%s:string = %s", k, v)
	}
	for k, v := range c.mapVals {
		fmt.Printf("\n%s:map = %v", k, v)
	}
	for k, v := range c.sliceVals {
		fmt.Printf("\n%s:slice = %v", k, v)
	}
	for k, v := range c.structVals {
		fmt.Printf("\n%s:struct = %v", k, v)
	}
}

func (c *Cfg) visit(cfgObj interface{}) error {
	queue := []*valueInfo{}
	queue = append(
		queue,
		&valueInfo{
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
		case k == reflect.Bool:
			c.boolVals[e.path] = e.v.Bool()
		case k == reflect.Int:
			c.intVals[e.path] = e.v.Interface().(int) // use int instead of uint/int/8/16/32/64
		case k == reflect.Float64:
			c.floatVals[e.path] = e.v.Float()
		case k == reflect.String:
			c.stringVals[e.path] = e.v.String()
		case k == reflect.Map:
			c.mapVals[e.path] = e.v.Interface()
		case k == reflect.Slice:
			sliceVal := e.v
			for i := 0; i < sliceVal.Len(); i++ {
				childName := fmt.Sprintf("%d", i)
				childValue := sliceVal.Index(i)
				childPath := fmt.Sprintf("%s[%s]", e.path, childName)
				if e.path == "" {
					// actually the root should not be an array
					childPath = childName
				}
				info := &valueInfo{
					v:    childValue,
					name: childName,
					path: childPath,
				}
				queue = append(queue, info)
			}
			// also set the whole slice as a config value
			c.sliceVals[e.path] = e.v.Interface()
		case k == reflect.Struct:
			structVal := e.v
			for i := 0; i < structVal.NumField(); i++ {
				childName := structVal.Type().Field(i).Name
				childValue := structVal.Field(i)
				childPath := fmt.Sprintf("%s.%s", e.path, childName)
				if e.path == "" {
					childPath = childName
				}

				// check if it should be retrieved from env
				tagValue := structVal.Type().Field(i).Tag.Get(GocfgTag)
				isEnv := strings.Contains(tagValue, GocfgValEnv)
				isRequired := strings.Contains(tagValue, GocfgValRequired)
				envName := strings.ToUpper(childName)

				if isEnv {
					envValue, exist := os.LookupEnv(envName)
					if !exist && isRequired {
						return fmt.Errorf("gocfg: warning: %s must be defined as an environment", envName)
					}
					fmt.Println(fmt.Sprintf("ENV.%s", envName), envValue)
					// set the value even it does not exist
					c.stringVals[fmt.Sprintf("ENV.%s", envName)] = envValue
					continue
				}

				info := &valueInfo{
					v:    childValue,
					name: childName,
					path: childPath,
				}
				queue = append(queue, info)
			}
			// also set the whole struct as a config value
			c.structVals[e.path] = e.v.Interface()
		case k == reflect.Ptr:
			info := &valueInfo{
				v:    e.v.Elem(),
				name: e.name,
				path: e.path,
			}
			queue = append(queue, info)
		case k == reflect.Invalid:
			if !e.v.IsValid() {
				c.warnf("gocfg: warning: %s(kind=%s) is zero value\n", e.path, k)
				// no op if the field is nil
				// From go doc:
				// IsValid reports whether v represents a value.
				// It returns false if v is the zero Value.
				// If IsValid returns false, all other methods except String panic.
				// Most functions and methods never return an invalid Value.
				// If one does, its documentation states the conditions explicitly.
			} else {
				// no op if it is zeroValue
				// Cfg will return zero value if this value is not set
				// therefore we don't set the value here
				return fmt.Errorf("gocfg: warning: %s(kind=%s) is invalid value", e.path, k)
			}
		default:
			return fmt.Errorf("gocfg: %s(kind=%s) is not supproted", e.path, k)
		}
	}

	return nil
}

// Bool get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) Bool(key string) (bool, bool) {
	val, ok := c.boolVals[key]
	return val, ok
}

// Int get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) Int(key string) (int, bool) {
	val, ok := c.intVals[key]
	return val, ok
}

// Float get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) Float(key string) (float64, bool) {
	val, ok := c.floatVals[key]
	return val, ok
}

// String get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) String(key string) (string, bool) {
	val, ok := c.stringVals[key]
	return val, ok
}

// Map get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) Map(key string) (interface{}, bool) {
	val, ok := c.mapVals[key]
	return val, ok
}

// Slice get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) Slice(key string) (interface{}, bool) {
	val, ok := c.sliceVals[key]
	return val, ok
}

// Struct get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) Struct(key string) (interface{}, bool) {
	val, ok := c.structVals[key]
	return val, ok
}

// BoolOr get a configuration value according to the key, or it returns the defaultVal instead.
func (c *Cfg) BoolOr(key string, defaultVal bool) bool {
	val, ok := c.boolVals[key]
	if ok {
		return val
	}
	return defaultVal
}

// IntOr get a configuration value according to the key, or it returns the defaultVal instead.
func (c *Cfg) IntOr(key string, defaultVal int) int {
	val, ok := c.intVals[key]
	if ok {
		return val
	}
	return defaultVal
}

// FloatOr get a configuration value according to the key, or it returns the defaultVal instead.
func (c *Cfg) FloatOr(key string, defaultVal float64) float64 {
	val, ok := c.floatVals[key]
	if ok {
		return val
	}
	return defaultVal
}

// StringOr get a configuration value according to the key, or it returns the defaultVal instead.
func (c *Cfg) StringOr(key string, defaultVal string) string {
	val, ok := c.stringVals[key]
	if ok {
		return val
	}
	return defaultVal
}

// MapOr get a configuration value according to the key, or it returns the defaultVal instead.
func (c *Cfg) MapOr(key string, defaultVal interface{}) interface{} {
	val, ok := c.mapVals[key]
	if ok {
		return val
	}
	return defaultVal
}

// SliceOr get a configuration value according to the key, or it returns the defaultVal instead.
func (c *Cfg) SliceOr(key string, defaultVal interface{}) interface{} {
	val, ok := c.sliceVals[key]
	if ok {
		return val
	}
	return defaultVal
}

// StructOr get a configuration value according to the key, or it returns the defaultVal instead.
func (c *Cfg) StructOr(key string, defaultVal interface{}) interface{} {
	val, ok := c.structVals[key]
	if ok {
		return val
	}
	return defaultVal
}

// GrabBool get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabBool(key string) bool { return c.boolVals[key] }

// GrabInt get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabInt(key string) int { return c.intVals[key] }

// GrabFloat get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabFloat(key string) float64 { return c.floatVals[key] }

// GrabString get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabString(key string) string { return c.stringVals[key] }

// GrabMap get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabMap(key string) interface{} { return c.mapVals[key] }

// GrabSlice get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabSlice(key string) interface{} { return c.sliceVals[key] }

// GrabStruct get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabStruct(key string) interface{} { return c.structVals[key] }

// SetBool set val in Cfg according to the key.
func (c *Cfg) SetBool(key string, val bool) { c.boolVals[key] = val }

// SetInt set val in Cfg according to the key.
func (c *Cfg) SetInt(key string, val int) { c.intVals[key] = val }

// SetFloat set val in Cfg according to the key.
func (c *Cfg) SetFloat(key string, val float64) { c.floatVals[key] = val }

// SetString set val in Cfg according to the key.
func (c *Cfg) SetString(key string, val string) { c.stringVals[key] = val }

// SetMap set val in Cfg according to the key.
func (c *Cfg) SetMap(key string, val interface{}) { c.mapVals[key] = val }

// SetSlice set val in Cfg according to the key.
func (c *Cfg) SetSlice(key string, val interface{}) { c.sliceVals[key] = val }

// SetStruct set val in Cfg according to the key.
func (c *Cfg) SetStruct(key string, val interface{}) { c.structVals[key] = val }
