package frecuenzy

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type frecuenzyManager struct {
	governors []string
}

func New() *frecuenzyManager {

	fm := &frecuenzyManager{
		governors: make([]string, 0),
	}
	err := fm.readGovernors()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return fm
}

func (fm *frecuenzyManager) readGovernors() error {

	cmdGetCpus := "ls /sys/devices/system/cpu/ | grep -cE '^cpu[0-9]*$'"
	output, err := exec.Command("bash", "-c", cmdGetCpus).Output()
	if err != nil {
		fmt.Println(err)
		return err
	}

	numCPUs, err := strconv.Atoi(strings.Replace(string(output), "\n", "", 1))
	for i := 0; i < numCPUs; i++ {
		file := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/scaling_governor", i)
		output, err := exec.Command("cat", file).Output()
		governor := strings.Replace(string(output), "\n", "", 1)
		fm.governors = append(fm.governors, governor)
		if err != nil {
			fmt.Println(err)
			return err
		}

	}
	return nil
}

func (fm *frecuenzyManager) Set(frequenzy int) error {
	freq := strconv.Itoa(frequenzy)
	cmd := exec.Command("cpupower", "frequency-set", "--freq", freq)
	err := cmd.Run()
	if err != nil {
		return err
	}
	cmd.Wait()
	return nil
}
func (fm *frecuenzyManager) Restore() error {
	for i, governor := range fm.governors {
		cmd := exec.Command("cpupower", "--cpu", strconv.Itoa(i), "frequency-set", "--governor", governor)
		err := cmd.Run()
		if err != nil {
			return err
		}
		cmd.Wait()
	}

	return nil
}
