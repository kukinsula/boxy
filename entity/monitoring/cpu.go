package monitoring

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	stat         = "/proc/stat"
	nbCpuColumns = 10
)

type CPU struct {
	LoadAverage     float64     `json:"average"`
	LoadAverages    []float64   `json:"averages"`
	CPUs            int         `json:"count"`
	CurrentMeasure  *CPUMeasure `json:"current"`
	previousMeasure *CPUMeasure
}

type CPUMeasure struct {
	CPUs              int   `json:"count"`
	SwitchContexts    int   `json:"switch-context"`
	BootTime          int64 `json:"boot-time"`
	Processes         int   `json:"processes"`
	ProcessorsRunning int   `json:"processors-running"`
	ProcessorsBlocked int   `json:"processors-blocked"`
	cpus              [][nbCpuColumns]int
}

func NewCPU() *CPU {
	cpus := runtime.NumCPU()

	return &CPU{
		CPUs:            cpus,
		CurrentMeasure:  newCPUMeasure(cpus),
		previousMeasure: newCPUMeasure(cpus),
		LoadAverages:    make([]float64, cpus),
	}
}

func newCPUMeasure(cpus int) *CPUMeasure {
	return &CPUMeasure{
		CPUs: cpus,
		cpus: make([][nbCpuColumns]int, cpus+1),
	}
}

func (cpu *CPU) Update() error {
	*cpu.previousMeasure = *cpu.CurrentMeasure
	copy((*cpu.previousMeasure).cpus, (*cpu.CurrentMeasure).cpus)
	cpu.CurrentMeasure.cpus = make([][nbCpuColumns]int, runtime.NumCPU()+1)

	err := cpu.CurrentMeasure.update()
	if err != nil {
		return err
	}

	cpu.computeCPUAverages()

	return nil
}

func (cpu *CPU) computeCPUAverages() {
	cpu.LoadAverage = cpu.computeCPULoad(cpu.CurrentMeasure.cpus[0], cpu.previousMeasure.cpus[0])

	for index := 0; index < cpu.CPUs; index++ {
		cpu.LoadAverages[index] = cpu.computeCPULoad(
			cpu.CurrentMeasure.cpus[index+1], cpu.previousMeasure.cpus[index+1])
	}
}

func (cpu *CPU) computeCPULoad(first, second [nbCpuColumns]int) float64 {
	numerator := float64((second[0] + second[1] + second[2]) -
		(first[0] + first[1] + first[2]))

	denominator := float64((second[0] + second[1] + second[2] + second[3]) -
		(first[0] + first[1] + first[2] + first[3]))

	return math.Abs(numerator / denominator * 100.0)
}

func (cpu *CPU) String() string {
	str := "\t========== CPU ==========\n\n"
	str += fmt.Sprintf("CPU: \t\t%.2f %%\n", cpu.LoadAverage)

	for index, average := range cpu.LoadAverages {
		str += fmt.Sprintf("CPU%d: \t\t%.2f %%\n", index, average)
	}

	str += fmt.Sprintf("\nSwitchContexts: \t\t%d\n",
		cpu.CurrentMeasure.SwitchContexts)

	str += fmt.Sprintf("BootTime: \t%d (%v)\n",
		cpu.CurrentMeasure.BootTime, time.Unix(cpu.CurrentMeasure.BootTime, 0))

	str += fmt.Sprintf("Processes: \t%d\n",
		cpu.CurrentMeasure.Processes)

	str += fmt.Sprintf("ProcessorsBlocked: \t%d\n",
		cpu.CurrentMeasure.ProcessorsBlocked)

	str += fmt.Sprintf("ProcessorsRunning: \t%d",
		cpu.CurrentMeasure.ProcessorsRunning)

	return str
}

func (measure *CPUMeasure) update() error {
	file, err := os.Open(stat)
	if err != nil {
		return err
	}
	defer file.Close()

	var lineName string
	var n, cpuCount int

	for scanner := bufio.NewScanner(file); scanner.Scan(); {
		line := scanner.Text()

		if strings.Contains(line, "cpu") {
			n, err = fmt.Sscanf(line, "%s %d %d %d %d %d %d %d %d %d %d", &lineName,
				&measure.cpus[cpuCount][0], &measure.cpus[cpuCount][1],
				&measure.cpus[cpuCount][2], &measure.cpus[cpuCount][3],
				&measure.cpus[cpuCount][4], &measure.cpus[cpuCount][5],
				&measure.cpus[cpuCount][6], &measure.cpus[cpuCount][7],
				&measure.cpus[cpuCount][8], &measure.cpus[cpuCount][9],
			)
			checkSscanf(lineName, err, n, 11)
			cpuCount++
		} else if strings.Contains(line, "ctxt") {
			n, err = fmt.Sscanf(line, "ctxt %d", &measure.SwitchContexts)
			checkSscanf("ctxt", err, n, 1)
		} else if strings.Contains(line, "btime") {
			n, err = fmt.Sscanf(line, "btime %d", &measure.BootTime)
			checkSscanf("btime", err, n, 1)
		} else if strings.Contains(line, "processes") {
			n, err = fmt.Sscanf(line, "processes %d", &measure.Processes)
			checkSscanf("processes", err, n, 1)
		} else if strings.Contains(line, "procs_running") {
			n, err = fmt.Sscanf(line, "procs_running %d", &measure.ProcessorsRunning)
			checkSscanf("procs_running", err, n, 1)
		} else if strings.Contains(line, "procs_blocked") {
			n, err = fmt.Sscanf(line, "procs_blocked %d", &measure.ProcessorsBlocked)
			checkSscanf("procs_blocked", err, n, 1)
		}
	}

	return nil
}

func (measure *CPUMeasure) String() string {
	var str string

	for index, data := range measure.cpus {
		str += fmt.Sprintf("CPU%d: %v\n", index, data)
	}

	str += fmt.Sprintf("SwitchContexts: %d\n", measure.SwitchContexts)

	return str
}
