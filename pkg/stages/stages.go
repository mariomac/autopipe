package stages

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/netobserv/gopipes/pkg/node"
	"github.com/sirupsen/logrus"
)

// A provider is a function that, given a configuration argument, returns a node with
// a function inside.
// If we implement this using Go 1.18 and generics, we could do the config argument as type safe.

type IngestProvider func(interface{}) *node.Init
type TransformProvider func(interface{}) *node.Middle
type ExportProvider func(interface{}) *node.Terminal

const defaultPort = 8080

type Http struct {
	Name string `hcl:",label"`
	Port int    `hcl:"port"`
}

// HttpIngestProvider listens for HTTP connections and forwards them
func HttpIngestProvider(cfg interface{}) *node.Init {
	c := cfg.(Http)
	port := c.Port
	if port == 0 {
		port = defaultPort
	}
	log := logrus.WithField("component", "HttpIngest")
	return node.AsInit(func(out chan<- []byte) {
		err := http.ListenAndServe(fmt.Sprintf(":%d", port),
			http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.Method != http.MethodPost {
					writer.WriteHeader(http.StatusBadRequest)
					return
				}
				body, err := ioutil.ReadAll(request.Body)
				if err != nil {
					log.WithError(err).Warn("failed request")
					writer.WriteHeader(http.StatusBadRequest)
					writer.Write([]byte(err.Error()))
					return
				}
				out <- body
			}))
		log.WithError(err).Warn("HTTP server ended")
	})
}

type Stdout struct {
	Name    string `hcl:",label"`
	Prepend string `hcl:"prepend"`
}

// StdOutExportProvider receives any message and prints it, prepending a given message
func StdOutExportProvider(cfg interface{}) *node.Terminal {
	c := cfg.(Stdout)
	return node.AsTerminal(func(in <-chan string) {
		for s := range in {
			fmt.Println(c.Prepend + s)
		}
	})
}

type Deleter struct {
	Name   string   `hcl:",label"`
	Fields []string `hcl:"fields"`
}

// FieldDeleterTransformProvider receives a map and removes the configured fields from it
func FieldDeleterTransformProvider(cfg interface{}) *node.Middle {
	c := cfg.(Deleter)
	toDelete := map[string]struct{}{}
	for _, f := range c.Fields {
		toDelete[fmt.Sprint(f)] = struct{}{}
	}
	return node.AsMiddle(func(in <-chan map[string]interface{}, out chan<- map[string]interface{}) {
		for m := range in {
			for td := range toDelete {
				delete(m, td)
			}
			out <- m
		}
	})
}