package gocfg

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestNormalCases(t *testing.T) {
	type config struct {
		BoolVal   bool    `json:"boolVal"`
		IntVal    int     `json:"intVal"`
		FloatVal  float64 `json:"floatVal"`
		StringVal string  `json:"stringVal" cfg:"env"`
		// MapVal    []*config `json:"mapVal"`
		SliceVal  []*config `json:"sliceVal"`
		StructVal *config   `json:"structVal"`
	}

	t.Run("test basic types", func(t *testing.T) {
		type Output struct {
			bools   map[string]bool
			ints    map[string]int
			floats  map[string]float64
			strings map[string]string
			maps    map[string]interface{}
			slices  map[string]interface{}
			structs map[string]interface{}
			envs    map[string]string
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

		envStringVal := "valueFromEnv"
		envs := map[string]string{
			"STRINGVAL": envStringVal,
		}
		for env, val := range envs {
			err := os.Setenv(env, val)
			if err != nil {
				t.Fatal(err)
			}
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
				envs: map[string]string{
					"ENV.STRINGVAL": envStringVal,
				},
			},
		}

		for i, input := range inputs {
			cfg, err := New(&config{}).Load(JSONStr(input))
			if err != nil {
				t.Fatal(err)
			}

			output := outputs[i]
			for key, val := range output.bools {
				if cfg.BoolOr(key, false) != val {
					t.Fatalf("key %s not match: expected: %t, got: %t", key, val, cfg.BoolOr(key, false))
				}
			}
			for key, val := range output.ints {
				if cfg.IntOr(key, -1) != val {
					t.Fatalf("key %s not match: expected: %d, got: %d", key, val, cfg.IntOr(key, -1))
				}
			}
			for key, val := range output.floats {
				if cfg.FloatOr(key, -1.0) != val {
					t.Fatalf("key %s not match: expected: %f, got: %f", key, val, cfg.FloatOr(key, -1.0))
				}
			}
			for key, val := range output.strings {
				if cfg.StringOr(key, "") != val {
					t.Fatalf("key %s not match: expected: %s, got: %s", key, val, cfg.StringOr(key, ""))
				}
			}
			for key, val := range output.envs {
				if cfg.StringOr(key, "") != val {
					t.Fatalf("key %s not match: expected: %s, got: %s", key, val, cfg.StringOr(key, ""))
				}
			}
		}
	})

	t.Run("overall tests", func(t *testing.T) {
		inputs := []int{
			5,
			8,
		}

		for _, configSize := range inputs {
			input, root := genConfig(configSize)
			cfg, err := New(&testConfig{}).Load(JSONStr(input))
			if err != nil {
				t.Fatal(err)
			}

			err = checkConfig(0, configSize, "", cfg, root)
			if err != nil {
				t.Fatal(err)
			}
		}
	})
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
	root := newNodeTree(0, size)
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

	if node.BoolVal != cfg.BoolOr(boolPath, false) {
		return fmt.Errorf("id:%d(%s) BoolVal not match %t %t", id, boolPath, node.BoolVal, cfg.BoolOr(boolPath, false))
	}
	if node.IntVal != cfg.IntOr(intPath, -1) {
		return fmt.Errorf("id:%d(%s) IntVal not match %d %d", id, intPath, node.IntVal, cfg.IntOr(intPath, -1))
	}
	if node.FloatVal != cfg.FloatOr(floatPath, -1.0) {
		return fmt.Errorf("id:%d(%s) FloatVal not match %f %f", id, floatPath, node.FloatVal, cfg.FloatOr(floatPath, -1.0))
	}
	if node.StringVal != cfg.StringOr(stringPath, "-1") {
		return fmt.Errorf("id:%d(%s) StringVal not match %s %s", id, stringPath, node.StringVal, cfg.StringOr(stringPath, "-1"))
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

// TODO: randomly add zero value
// although it should be same as getting non-existing value
func newNodeTree(id, size int) *testConfig {
	if id >= size {
		return nil
	}

	leftNode := newNodeTree(id+1, size)
	rightNode := newNodeTree(id+2, size)
	structNode := newNodeTree(id+3, size)
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
