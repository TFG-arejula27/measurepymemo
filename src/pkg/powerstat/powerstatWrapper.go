package powerstat

import (
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

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
	Power     float32
	Frecuency float32
}
type CStateData struct {
	Resident float32
	Count    int32
	Latency  int32
}

func Measure(time string) (PowerInfo, error) {
	pwrInf := PowerInfo{frames: make([]PowerInfoData, 0)}
	cmd := exec.Command("powerstat", "-R", "-c", "-z", "-n", "-f", "1", time)
	data, err := cmd.Output()
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

		} else if strings.HasPrefix("POLL", line) {
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

func getFrameInfoData(line string) (PowerInfoData, error) {
	parsedLine := strings.Split(line, " ")
	power, err := strconv.ParseFloat(parsedLine[9], 32)

	if err != nil {
		return PowerInfoData{}, err
	}
	frecuency, err := strconv.ParseFloat(parsedLine[10], 32)
	if err != nil {
		return PowerInfoData{}, err
	}

	return PowerInfoData{
		Power:     float32(power),
		Frecuency: float32(frecuency),
	}, nil
}

func getCstateData(line string) (CStateData, error) {
	var cstatedata = CStateData{}
	line = strings.Replace(line, "%", "", 1)
	parsedLine := strings.Split(line, " ")
	resident, err := strconv.ParseFloat(parsedLine[1], 32)
	if err != nil {
		return CStateData{}, err
	}
	cstatedata.Resident = float32(resident)
	if parsedLine[0] == "C0" {
		return cstatedata, nil
	}

	count, err := strconv.ParseFloat(parsedLine[2], 32)
	if err != nil {
		return CStateData{}, err
	}
	cstatedata.Count = int32(count)
	latency, err := strconv.ParseFloat(parsedLine[3], 32)
	if err != nil {
		return CStateData{}, err
	}
	cstatedata.Latency = int32(latency)
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
