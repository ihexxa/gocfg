package gocfg

import (
	"fmt"
	"reflect"
)

// TODO: add config size limit??

// ICfg is an interface defined for consumer according to *gocfg.Cfg
type ICfg interface {
	Bool(key string) (bool, bool)
	Int(key string) (int, bool)
	Float(key string) (float64, bool)
	String(key string) (string, bool)
	Map(key string) (interface{}, error)
	Slice(key string) (interface{}, error)
	Struct(key string) (interface{}, error)

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

	Load(URL string, config interface{}) error
	Print()
}

// Cfg is an abstraction over a configuration
type Cfg struct {
	BoolVals   map[string]bool
	IntVals    map[string]int
	FloatVals  map[string]float64
	StringVals map[string]string
	MapVals    map[string]interface{}
	SliceVals  map[string]interface{}
	StructVals map[string]interface{}
}

type valueInfo struct {
	v    reflect.Value
	path string
	name string
}

// New returns a new *Cfg
func New() *Cfg {
	return &Cfg{
		BoolVals:   map[string]bool{},
		IntVals:    map[string]int{},
		FloatVals:  map[string]float64{},
		StringVals: map[string]string{},
		MapVals:    map[string]interface{}{},
		SliceVals:  map[string]interface{}{},
		StructVals: map[string]interface{}{},
	}
}

// Load loads configuration from local path according to config's definition
func (c *Cfg) Load(pvd CfgProvider, config interface{}) error {
	err := pvd.Load(config)
	if err != nil {
		return err
	}
	return c.visit(config)
}

// Print prints all of the values in the Cfg
func (c *Cfg) Print() {
	for k, v := range c.BoolVals {
		fmt.Printf("\n%s:bool = %t", k, v)
	}
	for k, v := range c.IntVals {
		fmt.Printf("\n%s:int = %d", k, v)
	}
	for k, v := range c.FloatVals {
		fmt.Printf("\n%s:float = %f", k, v)
	}
	for k, v := range c.StringVals {
		fmt.Printf("\n%s:string = %s", k, v)
	}
	for k, v := range c.MapVals {
		fmt.Printf("\n%s:map = %v", k, v)
	}
	for k, v := range c.SliceVals {
		fmt.Printf("\n%s:slice = %v", k, v)
	}
	for k, v := range c.StructVals {
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
			c.BoolVals[e.path] = e.v.Bool()
		case k == reflect.Int:
			c.IntVals[e.path] = e.v.Interface().(int) // use int instead of uint/int/8/16/32/64
		case k == reflect.Float64:
			c.FloatVals[e.path] = e.v.Float()
		case k == reflect.String:
			c.StringVals[e.path] = e.v.String()
		case k == reflect.Map:
			c.MapVals[e.path] = e.v.Interface()
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
			c.SliceVals[e.path] = e.v.Interface()
		case k == reflect.Struct:
			structVal := e.v
			for i := 0; i < structVal.NumField(); i++ {
				childName := structVal.Type().Field(i).Name
				childValue := structVal.Field(i)
				childPath := fmt.Sprintf("%s.%s", e.path, childName)
				if e.path == "" {
					childPath = childName
				}
				info := &valueInfo{
					v:    childValue,
					name: childName,
					path: childPath,
				}
				queue = append(queue, info)
			}
			// also set the whole struct as a config value
			c.StructVals[e.path] = e.v.Interface()
		case k == reflect.Ptr:
			info := &valueInfo{
				v:    e.v.Elem(),
				name: e.name,
				path: e.path,
			}
			queue = append(queue, info)
		case k == reflect.Invalid:
			if !e.v.IsValid() {
				fmt.Printf("gocfg: warning: %s(kind=%s) is zero value\n", e.path, k)
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
	val, ok := c.BoolVals[key]
	return val, ok
}

// Int get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) Int(key string) (int, bool) {
	val, ok := c.IntVals[key]
	return val, ok
}

// Float get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) Float(key string) (float64, bool) {
	val, ok := c.FloatVals[key]
	return val, ok
}

// String get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) String(key string) (string, bool) {
	val, ok := c.StringVals[key]
	return val, ok
}

// Map get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) Map(key string) (interface{}, bool) {
	val, ok := c.MapVals[key]
	return val, ok
}

// Slice get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) Slice(key string) (interface{}, bool) {
	val, ok := c.SliceVals[key]
	return val, ok
}

// Struct get a configuration value according to key, the second returned value is false if nothing not found.
func (c *Cfg) Struct(key string) (interface{}, bool) {
	val, ok := c.StructVals[key]
	return val, ok
}

// GrabBool get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabBool(key string) bool { return c.BoolVals[key] }

// GrabInt get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabInt(key string) int { return c.IntVals[key] }

// GrabFloat get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabFloat(key string) float64 { return c.FloatVals[key] }

// GrabString get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabString(key string) string { return c.StringVals[key] }

// GrabMap get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabMap(key string) interface{} { return c.MapVals[key] }

// GrabSlice get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabSlice(key string) interface{} { return c.SliceVals[key] }

// GrabStruct get a configuration value according to the key, it returns zero value if no value is found.
func (c *Cfg) GrabStruct(key string) interface{} { return c.StructVals[key] }

// SetBool set val in Cfg according to the key.
func (c *Cfg) SetBool(key string, val bool) { c.BoolVals[key] = val }

// SetInt set val in Cfg according to the key.
func (c *Cfg) SetInt(key string, val int) { c.IntVals[key] = val }

// SetFloat set val in Cfg according to the key.
func (c *Cfg) SetFloat(key string, val float64) { c.FloatVals[key] = val }

// SetString set val in Cfg according to the key.
func (c *Cfg) SetString(key string, val string) { c.StringVals[key] = val }

// SetMap set val in Cfg according to the key.
func (c *Cfg) SetMap(key string, val interface{}) { c.MapVals[key] = val }

// SetSlice set val in Cfg according to the key.
func (c *Cfg) SetSlice(key string, val interface{}) { c.SliceVals[key] = val }

// SetStruct set val in Cfg according to the key.
func (c *Cfg) SetStruct(key string, val interface{}) { c.StructVals[key] = val }
