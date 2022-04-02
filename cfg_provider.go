package gocfg

import (
	"encoding/json"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v3"
)

// CfgProvider is a configuration loader interface
type CfgProvider interface {
	Load(dstCfg interface{}) error
}

// JSONStrCfg is a configuration loader for a json string
type JSONStrCfg struct {
	content string
}

// JSONStr inits a JSONStrCfg according to the content
func JSONStr(content string) *JSONStrCfg {
	return &JSONStrCfg{content: content}
}

// Load populates content according to the definition of the dstCfg
func (cfg *JSONStrCfg) Load(dstCfg interface{}) error {
	return json.Unmarshal([]byte(cfg.content), dstCfg)
}

// JSONCfg is a configuration loader for a local json file
type JSONCfg struct {
	path string
}

// JSON inits a JSONCfg according to the json file in the path
func JSON(path string) *JSONCfg {
	return &JSONCfg{path: path}
}

// Load populates json file according to the definition of the dstCfg
func (cfg *JSONCfg) Load(dstCfg interface{}) error {
	cfgFile, err := os.Open(cfg.path)
	if err != nil {
		return err
	}

	cfgBytes, err := ioutil.ReadAll(cfgFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(cfgBytes, dstCfg)
}

// YAMLCfg is a configuration loader for a local yaml file
type YAMLCfg struct {
	path string
}

// YAML inits a YAMLCfg according to the json file in the path
func YAML(path string) *YAMLCfg {
	return &YAMLCfg{path: path}
}

// Load populates yaml file according to the definition of the dstCfg
func (cfg *YAMLCfg) Load(dstCfg interface{}) error {
	cfgFile, err := os.Open(cfg.path)
	if err != nil {
		return err
	}

	cfgBytes, err := ioutil.ReadAll(cfgFile)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(cfgBytes, dstCfg)
}

// YAMLStrCfg is a configuration loader for a local yaml file
type YAMLStrCfg struct {
	content string
}

// YAMLStr inits a YAMLStrCfg according to the json file in the path
func YAMLStr(content string) *YAMLStrCfg {
	return &YAMLStrCfg{content: content}
}

// Load populates yaml file according to the definition of the dstCfg
func (cfg *YAMLStrCfg) Load(dstCfg interface{}) error {
	return yaml.Unmarshal([]byte(cfg.content), dstCfg)
}

// GoCfgCfg is a configuration loader for a gocfg struct
type GoCfgCfg struct {
	srcCfg *Cfg
}

// GoCfg inits a GoCfgCfg
func GoCfg(srcCfg *Cfg) *GoCfgCfg {
	return &GoCfgCfg{srcCfg: srcCfg}
}

// Load populates gocfg struct and save to
func (cfg *GoCfgCfg) Load(dstCfg interface{}) error {
	cfgBytes, err := json.Marshal(cfg.srcCfg.template)
	if err != nil {
		return err
	}
	return json.Unmarshal(cfgBytes, dstCfg)
}
