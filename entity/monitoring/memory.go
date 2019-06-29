package monitoring

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	meminfo = "/proc/meminfo"
)

type Memory struct {
	CurrentMeasure *MemoryMeasure
	lastMeasure    *MemoryMeasure

	// TODO: ajouter DeltaMemFree, DeltaMemOccupied, DeltaSwapFree, ...
}

type MemoryMeasure struct {
	MemTotal        kbyte `json:"total"`
	MemFree         kbyte `json:"free"`
	MemOccupied     kbyte `json:"occupied"`
	MemAvailable    kbyte `json:"available"`
	SwapTotal       kbyte `json:"swap-total"`
	SwapFree        kbyte `json:"swap-free"`
	SwapOccupied    kbyte `json:"swap-occupied"`
	VmallocTotal    kbyte `json:"vm-allocated-total"`
	VmallocFree     kbyte `json:"vm-allocated-free"`
	VmallocOccupied kbyte `json:"vm-allocated-occupied"`
}

func NewMemory() *Memory {
	return &Memory{
		CurrentMeasure: &MemoryMeasure{},
		lastMeasure:    &MemoryMeasure{},
	}
}

func (memory *Memory) Update() error {
	*memory.lastMeasure = *memory.CurrentMeasure

	return memory.CurrentMeasure.update()
}

func (memory *Memory) PercentMemFree() float64 {
	return 100.0 - memory.PercentMemOccupied()
}

func (memory *Memory) PercentMemOccupied() float64 {
	return float64(memory.CurrentMeasure.MemOccupied) * 100.0 /
		float64(memory.CurrentMeasure.MemTotal)
}

func (memory *Memory) PercentSwapFree() float64 {
	return 100.0 - memory.PercentSwapOccupied()
}

func (memory *Memory) PercentSwapOccupied() float64 {
	return float64(memory.CurrentMeasure.SwapOccupied) * 100.0 /
		float64(memory.CurrentMeasure.SwapTotal)
}

func (memory *Memory) PercentVmallocFree() float64 {
	return 100.0 - memory.PercentVmallocOccupied()
}

func (memory *Memory) PercentVmallocOccupied() float64 {
	return float64(memory.CurrentMeasure.VmallocOccupied) * 100.0 /
		float64(memory.CurrentMeasure.VmallocFree)
}

func (memory *Memory) String() string {
	format := "\t========== MEMORY ==========\n\n"
	format += "MemTotal:\t %s\n"
	format += "MemFree:\t %s\t%.3f %%\t(%s)\n"
	format += "MemOccupied:\t %s\t%.3f %%\t(%s)\n"
	format += "MemAvailable:\t %s\t\t\t(%s)\n"
	format += "SwapTotal:\t %s\n"
	format += "SwapFree:\t %s\t%.3f %%\t(%s)\n"
	format += "SwapOccupied:\t %s\t%.3f %%\t(%s)\n"
	format += "VmallocTotal:\t %s\n"
	format += "VmallocFree:\t %s\t%.3f %%\t(%s)\n"
	format += "VmallocOccupied: %s\t%.3f %%\t\t(%s)"

	return fmt.Sprintf(format,
		memory.CurrentMeasure.MemTotal,

		memory.CurrentMeasure.MemFree,
		memory.PercentMemFree(),
		memory.CurrentMeasure.MemFree-memory.lastMeasure.MemFree,

		memory.CurrentMeasure.MemOccupied,
		memory.PercentMemOccupied(),
		memory.CurrentMeasure.MemOccupied-memory.lastMeasure.MemOccupied,

		memory.CurrentMeasure.MemAvailable,
		memory.CurrentMeasure.MemAvailable-memory.lastMeasure.MemAvailable,

		memory.CurrentMeasure.SwapTotal,

		memory.CurrentMeasure.SwapFree,
		memory.PercentSwapFree(),
		memory.CurrentMeasure.SwapFree-memory.lastMeasure.SwapFree,

		memory.CurrentMeasure.SwapOccupied,
		memory.PercentSwapOccupied(),
		memory.CurrentMeasure.SwapOccupied-memory.lastMeasure.SwapOccupied,

		memory.CurrentMeasure.VmallocTotal,

		memory.CurrentMeasure.VmallocFree,
		memory.PercentVmallocFree(),
		memory.CurrentMeasure.VmallocFree-memory.lastMeasure.VmallocFree,

		memory.CurrentMeasure.VmallocOccupied,
		memory.PercentVmallocOccupied(),
		memory.CurrentMeasure.VmallocOccupied-memory.lastMeasure.VmallocOccupied)
}

func (measure *MemoryMeasure) update() error {
	file, err := os.Open(meminfo)
	if err != nil {
		return err
	}
	defer file.Close()

	var n int

	for scanner := bufio.NewScanner(file); scanner.Scan(); {
		line := scanner.Text()

		if strings.Contains(line, "MemTotal") {
			n, err = fmt.Sscanf(line, "MemTotal: %d kB", &measure.MemTotal)
			checkSscanf("MemTotal", err, n, 1)
		} else if strings.Contains(line, "MemFree") {
			n, err = fmt.Sscanf(line, "MemFree: %d kB", &measure.MemFree)
			checkSscanf("MemFree", err, n, 1)
		} else if strings.Contains(line, "MemAvailable") {
			n, err = fmt.Sscanf(line, "MemAvailable: %d kB", &measure.MemAvailable)
			checkSscanf("MemAvailable", err, n, 1)
		} else if strings.Contains(line, "SwapTotal") {
			n, err = fmt.Sscanf(line, "SwapTotal: %d kB", &measure.SwapTotal)
			checkSscanf("SwapTotal", err, n, 1)
		} else if strings.Contains(line, "SwapFree") {
			n, err = fmt.Sscanf(line, "SwapFree: %d kB", &measure.SwapFree)
			checkSscanf("SwapFree", err, n, 1)
		} else if strings.Contains(line, "VmallocTotal") {
			n, err = fmt.Sscanf(line, "VmallocTotal: %d kB", &measure.VmallocTotal)
			checkSscanf("VmallocTotal", err, n, 1)
		} else if strings.Contains(line, "VmallocUsed") {
			n, err = fmt.Sscanf(line, "VmallocUsed: %d kB", &measure.VmallocOccupied)
			checkSscanf("VmallocUsed", err, n, 1)
		}
	}

	measure.MemOccupied = measure.MemTotal - measure.MemFree
	measure.SwapOccupied = measure.SwapTotal - measure.SwapFree
	measure.VmallocFree = measure.VmallocTotal - measure.VmallocOccupied

	return nil
}
