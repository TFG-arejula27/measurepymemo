package powerstat

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

//measurer is a struct that allows external package to intercat with powerstat
type measurer struct {
	cmd *exec.Cmd
}

type PowerInfo struct {
	Message string        `json:"message"`
	Averge  PowerInfoData `json:"averge"`
	Max     PowerInfoData `json:"max"`
	Min     PowerInfoData `json:"min"`
	C1      CStateData    `json:"c1"`
	C2      CStateData    `json:"c2"`
	Poll    CStateData    `json:"poll"`
	C0      CStateData    `json:"c0"`
	frames  []PowerInfoData
	Time    time.Duration
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
	start := time.Now()
	data, err := m.cmd.Output()
	t := time.Now()
	pwrInf.Time = t.Sub(start)
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
		Power:     parsedLine[12],
		Frecuency: parsedLine[13],
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
	line += pwrInf.Message + ";"
	line += pwrInf.Averge.Power + ";" + pwrInf.Averge.Frecuency + ";"
	line += pwrInf.Max.Power + ";" + pwrInf.Max.Frecuency + ";"
	line += pwrInf.Min.Power + ";" + pwrInf.Min.Frecuency + ";"
	line += pwrInf.C2.Resident + ";" + pwrInf.C2.Count + ";" + pwrInf.C2.Latency + ";"
	line += pwrInf.C1.Resident + ";" + pwrInf.C1.Count + ";" + pwrInf.C1.Latency + ";"
	line += pwrInf.C0.Resident + ";" + pwrInf.C0.Count + ";" + pwrInf.C0.Latency + ";"
	line += pwrInf.Poll.Resident + ";" + pwrInf.Poll.Count + ";" + pwrInf.Poll.Latency + ";"
	line += strconv.FormatFloat(pwrInf.Time.Seconds(), 'f', 2, 64) + ";"
	line += "\n"
	return line
}

func (pwrInf *PowerInfo) GetCsvHeader() string {
	header := "Message;"
	header += "Average power(Watts);Average frecuenzy(GHz) ;"
	header += "Max power(Watts);Max frecuenzy(GHz) ;"
	header += "Min power(Watts);Min frecuenzy(GHz) ;"
	header += "C2 resident;C2 count;C2 latency;"
	header += "C1 resident;C1 count;C1 latency;"
	header += "C0 resident;C0 count;C0 latency;"
	header += "POLL resident;POLL count;POLL latency;"
	header += "time"
	header += "\n"
	return header
}

func (pwrInf *PowerInfo) GetHeader() string {
	header := fmt.Sprintf("%-20s%-20s%-20s%-20s", "Message", "Power(Watts)", "Frecuenzy(GHz)", "Time")
	header += "\n"
	return header
}

func (pwrInf *PowerInfo) GetData() string {
	time := strconv.FormatFloat(pwrInf.Time.Seconds(), 'f', 2, 64)
	line := fmt.Sprintf("%-20s%-20s%-20s%-20s", pwrInf.Message, pwrInf.Averge.Power, pwrInf.Averge.Frecuency, time)
	line += "\n"
	return line
}
