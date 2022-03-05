package main

import (
	"flag"
	"os"

	"github.com/mariomac/autopipe/pkg/config"
	"github.com/mariomac/autopipe/pkg/graph"
	"github.com/mariomac/autopipe/pkg/stage"
	"github.com/mariomac/autopipe/pkg/stage/system"
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
	builder.RegisterCodec(stage.BytesToStringCodec)
	builder.RegisterCodec(stage.JSONBytesToMapCodec)
	builder.RegisterCodec(stage.MapToStringCodec)

	// register the pipeline stages that are actually doing something
	builder.RegisterIngest(stage.HttpIngestProvider)
	builder.RegisterIngest(system.MonitorProvider)
	builder.RegisterTransform(stage.FieldDeleterTransformProvider)
	builder.RegisterExport(stage.StdOutExportProvider)

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
