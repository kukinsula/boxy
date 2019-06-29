package monitoring

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	dev          = "/proc/net/dev"
	nbNetColumns = 16
)

type Network struct {
	CurrentMeasure map[string]*NetworkInterface
	lastMeasures   map[string]*NetworkInterface
}

type NetworkInterface struct {
	Name     string  `json:"name"`
	Download float64 `json:"download"`
	Upload   float64 `json:"upload"`

	Measure [nbNetColumns]int64
}

func NewNetwork() *Network {
	return &Network{
		CurrentMeasure: make(map[string]*NetworkInterface),
		lastMeasures:   make(map[string]*NetworkInterface),
	}
}

func (network *Network) Update() error {
	network.lastMeasures = network.CurrentMeasure
	network.CurrentMeasure = make(map[string]*NetworkInterface)

	file, err := os.Open(dev)
	if err != nil {
		return err
	}
	defer file.Close()

	var data [nbNetColumns]int64
	var interfaceName string

	for scanner := bufio.NewScanner(file); scanner.Scan(); {
		line := scanner.Text()

		if strings.Contains(line, ":") {
			fields := strings.Fields(line)
			interfaceName = fields[0][:len(fields[0])-1]

			for i := 0; i < len(data); i++ {
				data[i], err = strconv.ParseInt(fields[i+1], 10, 0)
				if err != nil {
					log.Fatal(err)
				}
			}

			network.CurrentMeasure[interfaceName] = &NetworkInterface{
				interfaceName, 0.0, 0.0, data,
			}
		}
	}

	network.computeNetworkSpeed()

	return nil
}

func (network *Network) computeNetworkSpeed() {
	for name, _ := range network.CurrentMeasure {
		if network.lastMeasures[name] != nil {
			network.CurrentMeasure[name].Download =
				float64(network.CurrentMeasure[name].Measure[0]-network.lastMeasures[name].Measure[0]) / float64(1000000)

			network.CurrentMeasure[name].Upload =
				float64(network.CurrentMeasure[name].Measure[9]-network.lastMeasures[name].Measure[9]) / float64(1000000)
		}
	}
}

func (network *Network) String() string {
	str := "\t========== NETWORK ==========\n\n"

	for _, net := range network.CurrentMeasure {
		str += fmt.Sprintf("%s:\tDownload: %f MB/s,\tUpload: %f MB/s\n",
			net.Name, net.Download, net.Upload)
	}

	return str
}

func isInterface(str string) bool {
	valid := false

	switch str {
	case "wlan0", "l0":
		valid = true
	}

	return valid
}
