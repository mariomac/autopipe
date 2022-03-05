// Package system ingests monitoring information from system
package system

import (
	"fmt"
	"time"

	"github.com/mariomac/autopipe/pkg/stage"
	"github.com/netobserv/gopipes/pkg/node"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/sirupsen/logrus"
)

const defaultInterval = 5 * time.Second

type Monitor struct {
	Name     string `hcl:",label"`
	Interval string `hcl:"interval,optional"`
}

var MonitorProvider = stage.IngestProvider{
	StageType: "sysmon",
	Instantiator: func(cfg interface{}) *node.Init {
		log := logrus.WithField("stageType", "sysmon")
		interval, err := time.ParseDuration(cfg.(Monitor).Interval)
		if err != nil {
			log.WithField("interval", defaultInterval).Infof("using default interval")
			interval = defaultInterval
		}
		return node.AsInit(func(out chan<- map[string]interface{}) {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			// TODO: add a feature to gopipes to allow stopping nodes
			for {
				<-ticker.C
				stats := map[string]interface{}{}
				vmem, err := mem.VirtualMemory()
				if err != nil {
					log.WithError(err).Warn("can't fetch virtual memory. Skipping")
					continue
				}
				stats["memoryPercent"] = 100 * vmem.Available / vmem.Total
				scpu, err := cpu.Percent(0, true)
				if err != nil {
					log.WithError(err).Warn("can't fetch cpu stats. Skipping")
					continue
				}
				for i, sc := range scpu {
					stats[fmt.Sprintf("cpuPercent%d", i)] = sc
				}
				scpu, err = cpu.Percent(0, false)
				if err != nil {
					log.WithError(err).Warn("can't fetch cpu stats. Skipping")
					continue
				}
				stats["cpuPercentTotal"] = scpu[0]
				nios, err := net.IOCounters(false)
				if err != nil {
					log.WithError(err).Warn("can't fetch cpu stats. Skipping")
					continue
				}
				for _, nio := range nios {
					stats["netBytesSentPerSecond_"+nio.Name] = nio.BytesSent
					stats["netBytesRecvPerSecond_"+nio.Name] = nio.BytesRecv
				}
				out <- stats
			}
		})
	},
}
