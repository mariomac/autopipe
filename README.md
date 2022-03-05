# Experiment: automatic connection of pipeline nodes

Layer on top of [GoPipes](https://github.com/netobserv/gopipes) library that allows
a user building its own processing pipeline/graph by choosing a subset of
predefined nodes and interconnecting them.

The main points this proof of concept are trying to validate are:

* Automatic insertion of encoder/decoders between pipeline stages that have
  incompatible output->input types. Each codec needs to be pre-registered
  in the system for any possible output->input connection.
* Usage of [HCL for defining the pipeline](cmd/nodes.hcl) nodes and their connections, as a
  human-friendlier alternative to YAML.
* A Functional approach to instantiate and configure each pipeline stage, simplifying and
  unifying the registration of pipe stages.

The provided [example](./cmd) ingests HTTP JSONs, removes some "prohibited" fields, and prints the JSON in the standard output. It also prints the original JSON for debugging purposes.

Running the example:

```
go run cmd/autopipe.go -graph cmd/nodes.hcl
```

Then, you can submit some JSONs to the running ingester:

```
curl -X POST -d '{"hello":"my friend","password":"sup3rs3cr37","secret":"kadlfjjsdlaf"}' http://localhost:8080
```

You will see some output in the "autopipe" command log:

```
Received message: {"hello":"my friend","password":"sup3rs3cr37","secret":"kadlfjjsdlaf"}
Safe-to-show message: {"hello":"my friend"}
```

For an explanation on what's happening, read the [cmd/nodes.hcl](cmd/nodes.hcl) file.
