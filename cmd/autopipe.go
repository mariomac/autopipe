package main

import (
	"flag"
	"os"

	"github.com/mariomac/autopipe/pkg/config"
	"github.com/mariomac/autopipe/pkg/graph"
	"github.com/mariomac/autopipe/pkg/stages"
	"github.com/sirupsen/logrus"
)

var graphFile = flag.String("graph", "", "HCL graph file")

func main() {
	flag.Parse()
	if graphFile == nil || *graphFile == "" {
		flag.PrintDefaults()
		os.Exit(-1)
	}
	builder := graph.NewBuilder()

	// register codecs for automatic transformation between incompatible stages
	builder.RegisterCodec(stages.BytesToStringCodec)
	builder.RegisterCodec(stages.JSONBytesToMapCodec)
	builder.RegisterCodec(stages.MapToStringCodec)

	// register the pipeline stages that are actually doing something
	builder.RegisterIngest("http", stages.HttpIngestProvider)
	builder.RegisterTransform("deleter", stages.FieldDeleterTransformProvider)
	builder.RegisterExport("stdout", stages.StdOutExportProvider)

	// Parse config and build graph from it
	grp, err := os.Open(*graphFile)
	if err != nil {
		logrus.WithError(err).Fatal("can't load configuration")
	}
	cfg, err := config.ReadConfig(grp)
	if err != nil {
		logrus.WithError(err).Fatal("can't instantiate configuration")
	}
	config.ApplyConfig(&cfg, builder)

	// build and run the graph
	b := builder.Build()
	b.Run()
}
