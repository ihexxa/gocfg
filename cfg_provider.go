package gocfg

import (
	"encoding/json"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v3"
)

// URLSrcProvider is a configuration loader interface
type CfgProvider interface {
	Load(dstCfg interface{}) error
}

type JSONStrCfg struct {
	content string
}

func JSONStr(content string) *JSONStrCfg {
	return &JSONStrCfg{content: content}
}

func (cfg *JSONStrCfg) Load(dstCfg interface{}) error {
	return json.Unmarshal([]byte(cfg.content), dstCfg)
}

type JSONCfg struct {
	path string
}

func JSON(path string) *JSONCfg {
	return &JSONCfg{path: path}
}

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

type YAMLCfg struct {
	path string
}

func YAML(path string) *YAMLCfg {
	return &YAMLCfg{path: path}
}

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
