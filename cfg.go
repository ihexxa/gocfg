package gocfg

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
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

	Bools() map[string]bool
	Ints() map[string]int
	Floats() map[string]float64
	Strings() map[string]string
	Maps() map[string]interface{}
	Slices() map[string]interface{}
	Structs() map[string]interface{}

	SetBool(key string, val bool)
	SetInt(key string, val int)
	SetFloat(key string, val float64)
	SetString(key string, val string)
	SetStruct(key string, val interface{})

	Print()
	Debug()
	ToString() string
	JSON() (string, error)
	Template() interface{}
}

// Cfg is an abstraction over a configuration
type Cfg struct {
	debug    bool
	template interface{}

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
func New(template interface{}) *Cfg {
	return &Cfg{
		debug:      false,
		template:   template,
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
func (c *Cfg) Load(pvds ...CfgProvider) (*Cfg, error) {
	var err error
	for _, pvd := range pvds {
		if err = pvd.Load(c.template); err != nil {
			return nil, err
		}
		if err = c.visit(c.template); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Cfg) warnf(format string, vals ...interface{}) {
	if c.debug {
		fmt.Printf(format, vals...)
	}
}

// JSON returns all configs as a JSON in a string
func (c *Cfg) JSON() (string, error) {
	tpltBytes, err := json.Marshal(c.template)
	if err != nil {
		return "", err
	}
	return string(tpltBytes), nil
}

// Template returns all configs as an interface{}
func (c *Cfg) Template() interface{} {
	return c.template
}

// Print prints all of the values in the Cfg
func (c *Cfg) Print() {
	fmt.Println(c.ToString())
}

// String returns all of the values in the Cfg as a string
func (c *Cfg) ToString() string {
	keys := []string{}
	rows := []string{}

	for k := range c.boolVals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := c.boolVals[k]
		rows = append(rows, fmt.Sprintf("%s:bool = %t", k, v))
	}
	keys = keys[:0]

	for k := range c.intVals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := c.intVals[k]
		rows = append(rows, fmt.Sprintf("%s:int = %d", k, v))
	}
	keys = keys[:0]

	for k := range c.floatVals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := c.floatVals[k]
		rows = append(rows, fmt.Sprintf("%s:float = %f", k, v))
	}
	keys = keys[:0]

	for k := range c.stringVals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := c.stringVals[k]
		rows = append(rows, fmt.Sprintf("%s:string = %s", k, v))
	}
	keys = keys[:0]

	for k := range c.mapVals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		mv := c.mapVals[k]
		mvBytes, err := json.Marshal(mv)
		if err != nil {
			panic(err)
		}
		rows = append(rows, fmt.Sprintf("%s:map = %v", k, string(mvBytes)))
	}
	keys = keys[:0]

	for k := range c.sliceVals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sv := c.sliceVals[k]
		svBytes, err := json.Marshal(sv)
		if err != nil {
			panic(err)
		}
		rows = append(rows, fmt.Sprintf("%s:slice = %v", k, string(svBytes)))
	}
	keys = keys[:0]

	for k := range c.structVals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sv := c.structVals[k]
		svBytes, err := json.Marshal(sv)
		if err != nil {
			panic(err)
		}
		rows = append(rows, fmt.Sprintf("%s:struct = %v", k, string(svBytes)))
	}
	keys = keys[:0]

	return strings.Join(rows, "\n")
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
					// set the value even it does not exist
					c.stringVals[fmt.Sprintf("ENV.%s", envName)] = envValue
				}

				// also set the config value accodingly
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

func (c *Cfg) Bools() map[string]bool {
	return c.boolVals
}

func (c *Cfg) Ints() map[string]int {
	return c.intVals
}

func (c *Cfg) Floats() map[string]float64 {
	return c.floatVals
}

func (c *Cfg) Strings() map[string]string {
	return c.stringVals
}

func (c *Cfg) Maps() map[string]interface{} {
	return c.mapVals
}

func (c *Cfg) Slices() map[string]interface{} {
	return c.sliceVals
}

func (c *Cfg) Structs() map[string]interface{} {
	return c.structVals
}
