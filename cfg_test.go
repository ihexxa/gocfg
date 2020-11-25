package gocfg

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNormalCases(t *testing.T) {
	// t.Run("normal cases", testBaiscCases)
	t.Run("overall tests", testStructCases)
}

func testBaiscCases(t *testing.T) {
	type Output struct {
		bools   map[string]bool
		ints    map[string]int
		floats  map[string]float64
		strings map[string]string
		maps    map[string]interface{}
		slices  map[string]interface{}
		structs map[string]interface{}
	}

	type config struct {
		BoolVal   bool    `json:"boolVal"`
		IntVal    int     `json:"intVal"`
		FloatVal  float64 `json:"floatVal"`
		StringVal string  `json:"stringVal"`
		// MapVal    []*config `json:"mapVal"`
		SliceVal  []*config `json:"sliceVal"`
		StructVal *config   `json:"structVal"`
	}

	inputs := []string{
		`
		{
			"boolVal": true,
			"intVal": 1,
			"floatVal": 1,
			"stringVal": "1",
			"sliceVal": [
				{
					"boolVal": false,
					"intVal": 11,
					"floatVal": 1.1,
					"stringVal": "11",
					"sliceVal": null,
					"structVal": null
				},
				{
					"boolVal": false,
					"intVal": 12,
					"floatVal": 1.2,
					"stringVal": "12",
					"sliceVal": [],
					"structVal": {}
				}
			],
			"structVal": {
				"boolVal": false,
				"intVal": 2,
				"floatVal": 2.0,
				"stringVal": "2",
				"sliceVal": [],
				"structVal": {}
			}
		}
		`,
	}

	outputs := []*Output{
		&Output{
			bools: map[string]bool{
				"BoolVal":             true,
				"SliceVal[0].boolVal": false,
				"SliceVal[1].boolVal": false,
				"StructVal.BoolVal":   false,
			},
			ints: map[string]int{
				"IntVal":             1,
				"SliceVal[0].IntVal": 11,
				"SliceVal[1].IntVal": 12,
				"StructVal.IntVal":   2,
			},
			floats: map[string]float64{
				"FloatVal":             1.0,
				"SliceVal[0].FloatVal": 1.1,
				"SliceVal[1].FloatVal": 1.2,
				"StructVal.FloatVal":   2.0,
			},
			strings: map[string]string{
				"StringVal":             "1",
				"SliceVal[0].StringVal": "11",
				"SliceVal[1].StringVal": "12",
				"StructVal.StringVal":   "2",
			},
		},
	}

	for i, input := range inputs {
		cfg := New()
		err := cfg.Load(JSONStr(input), &config{})
		if err != nil {
			t.Fatal(err)
		}

		output := outputs[i]
		for key, val := range output.bools {
			if cfg.GrabBool(key) != val {
				t.Fatalf("key %s not match: expected: %t, got: %t", key, val, cfg.GrabBool(key))
			}
		}
		for key, val := range output.ints {
			if cfg.GrabInt(key) != val {
				t.Fatalf("key %s not match: expected: %d, got: %d", key, val, cfg.GrabInt(key))
			}
		}
		for key, val := range output.floats {
			if cfg.GrabFloat(key) != val {
				t.Fatalf("key %s not match: expected: %f, got: %f", key, val, cfg.GrabFloat(key))
			}
		}
		for key, val := range output.strings {
			if cfg.GrabString(key) != val {
				t.Fatalf("key %s not match: expected: %s, got: %s", key, val, cfg.GrabString(key))
			}
		}
	}
}

type testConfig struct {
	BoolVal   bool          `json:"boolVal"`
	IntVal    int           `json:"intVal"`
	FloatVal  float64       `json:"floatVal"`
	StringVal string        `json:"stringVal"`
	SliceVal  []*testConfig `json:"sliceVal"`
	StructVal *testConfig   `json:"structVal"`
}

func genConfig(size int) (string, *testConfig) {
	root := newCfgNode(0, size)
	cfgBytes, err := json.Marshal(root)
	if err != nil {
		panic(err) // TODO: use Fatalf instead
	}

	return string(cfgBytes), root
}

func checkConfig(id, size int, path string, cfg *Cfg, node *testConfig) error {
	boolPath := fmt.Sprintf("%s.BoolVal", path)
	intPath := fmt.Sprintf("%s.IntVal", path)
	floatPath := fmt.Sprintf("%s.FloatVal", path)
	stringPath := fmt.Sprintf("%s.StringVal", path)
	if path == "" {
		// for direct children of the root, there no dot in the path
		boolPath = "BoolVal"
		intPath = "IntVal"
		floatPath = "FloatVal"
		stringPath = "StringVal"
	}

	if node.BoolVal != cfg.GrabBool(boolPath) {
		return fmt.Errorf("id:%d BoolVal not match %t %t %s", id, node.BoolVal, cfg.GrabBool(boolPath), boolPath)
	}
	if node.IntVal != cfg.GrabInt(intPath) {
		return fmt.Errorf("id:%d IntVal not match %d %d %s", id, node.IntVal, cfg.GrabInt(intPath), intPath)
	}
	if node.FloatVal != cfg.GrabFloat(floatPath) {
		return fmt.Errorf("id:%d FloatVal not match %f %f %s", id, node.FloatVal, cfg.GrabFloat(floatPath), floatPath)
	}
	if node.StringVal != cfg.GrabString(stringPath) {
		return fmt.Errorf("id:%d StringVal not match %s %s %s", id, node.StringVal, cfg.GrabString(stringPath), stringPath)
	}

	prefix := fmt.Sprintf("%s.", path)
	if path == "" {
		prefix = ""
	}

	for _, i := range []int{1, 2} {
		if id+i >= size {
			return nil
		} else if len(node.SliceVal) < i {
			return fmt.Errorf("id:%d incorrect slice size: %v", id, node)
		} else {
			err := checkConfig(id+i, size, fmt.Sprintf("%sSliceVal[%d]", prefix, i-1), cfg, node.SliceVal[i-1])
			if err != nil {
				return err
			}
		}
	}

	if id+3 >= size {
		return nil
	} else if node.StructVal == nil {
		return fmt.Errorf("id:%d struct should not be nil: %v", id, node)
	} else if err := checkConfig(id+3, size, fmt.Sprintf("%sStructVal", prefix), cfg, node.StructVal); err != nil {
		return err
	}

	return nil
}

func newCfgNode(id, size int) *testConfig {
	if id >= size {
		// TODO: randomly add zero value
		// although it should be same in getting value
		return nil
	}

	leftNode := newCfgNode(id+1, size)
	rightNode := newCfgNode(id+2, size)
	structNode := newCfgNode(id+3, size)
	return &testConfig{
		BoolVal:   id%2 == 0,
		IntVal:    id,
		FloatVal:  float64(id),
		StringVal: fmt.Sprintf("%d", id),
		SliceVal: []*testConfig{
			leftNode,
			rightNode,
		},
		StructVal: structNode,
	}
}

func testStructCases(t *testing.T) {
	inputs := []int{
		5,
		8,
	}

	var err error
	for _, configSize := range inputs {
		input, root := genConfig(configSize)
		cfg := New()
		err = cfg.Load(JSONStr(input), &testConfig{})
		if err != nil {
			t.Fatal(err)
		}

		err = checkConfig(0, configSize, "", cfg, root)
		if err != nil {
			t.Fatal(err)
		}
	}
}
