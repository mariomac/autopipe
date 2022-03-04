package config

import (
	"io"
	"io/ioutil"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mariomac/autopipe/pkg/graph"
	"github.com/mariomac/autopipe/pkg/stages"
	"github.com/sirupsen/logrus"
)

type PipeConfig struct {
	Http    []stages.Http    `hcl:"http,block"`
	StdOut  []stages.Stdout  `hcl:"stdout,block"`
	Deleter []stages.Deleter `hcl:"deleter,block"`
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
	// TODO: probably it's a better way to configure HCL to not have to iterate by all the stage types
	for _, stg := range cfg.StdOut {
		if err := builder.Instantiate(graph.NodeName(stg.Name), "stdout", stg); err != nil {
			logrus.WithError(err).WithField("config", stg).Fatal("can't instantiate node")
		}
	}
	for _, stg := range cfg.Http {
		if err := builder.Instantiate(graph.NodeName(stg.Name), "http", stg); err != nil {
			logrus.WithError(err).WithField("config", stg).Fatal("can't instantiate node")
		}
	}
	for _, stg := range cfg.Deleter {
		if err := builder.Instantiate(graph.NodeName(stg.Name), "deleter", stg); err != nil {
			logrus.WithError(err).WithField("config", stg).Fatal("can't instantiate node")
		}
	}
	for src, dsts := range cfg.Connect {
		for _, dst := range dsts {
			if err := builder.Connect(graph.NodeName(src), graph.NodeName(dst)); err != nil {
				logrus.WithError(err).
					WithFields(logrus.Fields{"src": src, "dst": dst}).
					Fatal("can't connect stages")
			}
		}
	}
}
