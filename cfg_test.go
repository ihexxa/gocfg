package gocfg

import "testing"

func TestNormalCases(t *testing.T) {
	t.Run("normal cases", testBaiscCases)
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
		cfger := New()
		err := cfger.Load(JSONStr(input), &config{})
		if err != nil {
			t.Fatal(err)
		}

		output := outputs[i]
		for key, val := range output.bools {
			if cfger.GrabBool(key) != val {
				t.Fatalf("key %s not match: expected: %t, got: %t", key, val, cfger.GrabBool(key))
			}
		}
		for key, val := range output.ints {
			if cfger.GrabInt(key) != val {
				t.Fatalf("key %s not match: expected: %d, got: %d", key, val, cfger.GrabInt(key))
			}
		}
		for key, val := range output.floats {
			if cfger.GrabFloat(key) != val {
				t.Fatalf("key %s not match: expected: %f, got: %f", key, val, cfger.GrabFloat(key))
			}
		}
		for key, val := range output.strings {
			if cfger.GrabString(key) != val {
				t.Fatalf("key %s not match: expected: %s, got: %s", key, val, cfger.GrabString(key))
			}
		}
	}
}
