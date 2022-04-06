package powerstat

import (
	"log"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

//measurer is a struct that allows external package to intercat with powerstat
type measurer struct {
	cmd *exec.Cmd
}

type PowerInfo struct {
	Averge PowerInfoData `json:"averge"`
	Max    PowerInfoData `json:"max"`
	Min    PowerInfoData `json:"min"`
	C1     CStateData    `json:"c1"`
	C2     CStateData    `json:"c2"`
	Poll   CStateData    `json:"poll"`
	C0     CStateData    `json:"c0"`
	frames []PowerInfoData
}
type PowerInfoData struct {
	Power     string
	Frecuency string
}
type CStateData struct {
	Resident string
	Count    string
	Latency  string
}

//New creates a struct measurer, initializing the value of time to one minute
func New(time string) *measurer {

	return &measurer{
		cmd: exec.Command("powerstat", "-R", "-c", "-z", "-n", "-f", "1", time),
	}
}

//Run executes powerstate
func (m *measurer) Run() (PowerInfo, error) {
	pwrInf := PowerInfo{frames: make([]PowerInfoData, 0)}
	data, err := m.cmd.Output()
	if err != nil {
		return pwrInf, err

	}
	output := string(data)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")

		if isEmptyLine(line) {
			continue
		}
		if isFrame(line) {

			frame, err := getFrameInfoData(line)
			if err != nil {
				log.Println("cant get frame info")
				continue
			}
			pwrInf.frames = append(pwrInf.frames, frame)

		} else if strings.HasPrefix(line, "Average") {
			frameInfoData, err := getFrameInfoData(line)
			if err != nil {
				log.Println("cant get average info")
				continue
			}
			pwrInf.Averge = frameInfoData

		} else if strings.HasPrefix(line, "Minimum") {
			frameInfoData, err := getFrameInfoData(line)
			if err != nil {
				log.Println("cant get minimun info")
				continue
			}
			pwrInf.Min = frameInfoData

		} else if strings.HasPrefix(line, "Maximum") {
			frameInfoData, err := getFrameInfoData(line)
			if err != nil {
				log.Println("cant get maximun info")
				continue
			}
			pwrInf.Max = frameInfoData

		} else if strings.HasPrefix(line, "C2") {
			cstatedata, err := getCstateData(line)
			if err != nil {
				log.Println("cant get minimun C2 state info")
				continue
			}
			pwrInf.C2 = cstatedata

		} else if strings.HasPrefix(line, "C1") {
			cstatedata, err := getCstateData(line)
			if err != nil {
				log.Println("cant get minimun C1 state info")
				continue
			}
			pwrInf.C1 = cstatedata

		} else if strings.HasPrefix(line, "C0") {
			cstatedata, err := getCstateData(line)
			if err != nil {
				log.Println("cant get minimun C0 state info")
				continue
			}
			pwrInf.C0 = cstatedata

		} else if strings.HasPrefix(line, "POLL") {
			cstatedata, err := getCstateData(line)
			if err != nil {
				log.Println("cant get minimun POLL state info")
				continue
			}
			pwrInf.Poll = cstatedata

		} else {
			continue
		}
	}
	//parsedLines = append(parsedLines, strings.Split(line, " "))

	return pwrInf, nil

}

func (m *measurer) End() {
	m.cmd.Process.Signal(syscall.SIGINT)
}

func getFrameInfoData(line string) (PowerInfoData, error) {
	parsedLine := strings.Split(line, " ")

	return PowerInfoData{
		Power:     parsedLine[9],
		Frecuency: parsedLine[10],
	}, nil
}

func getCstateData(line string) (CStateData, error) {
	var cstatedata = CStateData{}
	line = strings.Replace(line, "%", "", 1)
	parsedLine := strings.Split(line, " ")
	resident := parsedLine[1]
	cstatedata.Resident = resident
	if parsedLine[0] == "C0" {
		return cstatedata, nil
	}

	cstatedata.Count = parsedLine[2]
	cstatedata.Latency = parsedLine[3]

	return cstatedata, nil

}

func isFrame(line string) bool {
	res := regexp.MustCompile(`^[0-9\.].*$`).MatchString(line)
	return res

}
func framesEnded(line string) bool {
	res := regexp.MustCompile(`^[ ]*Average`).MatchString(line)
	return res

}
func isEmptyLine(line string) bool {
	return line == ""
}

func (pwrInf *PowerInfo) ToCsv() string {
	line := ""
	line += pwrInf.Averge.Power + ";" + pwrInf.Averge.Frecuency + ";"
	line += pwrInf.Max.Power + ";" + pwrInf.Max.Frecuency + ";"
	line += pwrInf.Min.Power + ";" + pwrInf.Min.Frecuency + ";"
	line += pwrInf.C2.Resident + ";" + pwrInf.C2.Count + ";" + pwrInf.C2.Latency + ";"
	line += pwrInf.C1.Resident + ";" + pwrInf.C1.Count + ";" + pwrInf.C1.Latency + ";"
	line += pwrInf.C0.Resident + ";" + pwrInf.C0.Count + ";" + pwrInf.C0.Latency + ";"
	line += pwrInf.Poll.Resident + ";" + pwrInf.Poll.Count + ";" + pwrInf.Poll.Latency + ";"
	line += "\n"
	return line
}
