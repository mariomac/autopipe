package config

import (
	"io"
	"io/ioutil"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mariomac/autopipe/pkg/graph"
	"github.com/mariomac/autopipe/pkg/stage"
	"github.com/mariomac/autopipe/pkg/stage/system"
	"github.com/sirupsen/logrus"
)

type PipeConfig struct {
	Http    []stage.Http     `hcl:"http,block"`
	StdOut  []stage.Stdout   `hcl:"stdout,block"`
	Deleter []stage.Deleter  `hcl:"deleter,block"`
	SysMon  []system.Monitor `hcl:"sysmon,block"`
	Connect Connections      `hcl:"connect"`
}

// Connections key: name of the source node. Value: array of destination nodes.
type Connections map[string][]string

func ReadConfig(in io.Reader) (PipeConfig, error) {
	src, err := ioutil.ReadAll(in)
	if err != nil {
		return PipeConfig{}, err
	}
	var pc PipeConfig
	err = hclsimple.Decode(".hcl", src, nil, &pc)
	return pc, err
}

func ApplyConfig(cfg *PipeConfig, builder *graph.Builder) {
	// TODO: find a better way to configure from HCL without having to iterate all the stage types
	for _, stg := range cfg.StdOut {
		if err := builder.Instantiate(stage.Name(stg.Name), stage.StdOutExportProvider.StageType, stg); err != nil {
			logrus.WithError(err).WithField("config", stg).Fatal("can't instantiate node")
		}
	}
	for _, stg := range cfg.Http {
		if err := builder.Instantiate(stage.Name(stg.Name), stage.HttpIngestProvider.StageType, stg); err != nil {
			logrus.WithError(err).WithField("config", stg).Fatal("can't instantiate node")
		}
	}
	for _, stg := range cfg.Deleter {
		if err := builder.Instantiate(stage.Name(stg.Name), stage.FieldDeleterTransformProvider.StageType, stg); err != nil {
			logrus.WithError(err).WithField("config", stg).Fatal("can't instantiate node")
		}
	}
	for _, stg := range cfg.SysMon {
		if err := builder.Instantiate(stage.Name(stg.Name), system.MonitorProvider.StageType, stg); err != nil {
			logrus.WithError(err).WithField("config", stg).Fatal("can't instantiate node")
		}
	}
	for src, dsts := range cfg.Connect {
		for _, dst := range dsts {
			if err := builder.Connect(stage.Name(src), stage.Name(dst)); err != nil {
				logrus.WithError(err).
					WithFields(logrus.Fields{"src": src, "dst": dst}).
					Fatal("can't connect stages")
			}
		}
	}
}
