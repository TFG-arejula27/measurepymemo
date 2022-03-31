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
	C1     Cstate        `json:"c1"`
	C2     Cstate        `json:"c2"`
	Poll   float32       `json:"poll"`
	C0     Cstate        `json:"c0"`
	frames []PowerInfoData
}
type PowerInfoData struct {
	Power     float32
	Frecuency float32
}
type Cstate struct {
	Resident float32
	Count    int32
	Latency  int32
}

func Measure() error {
	pwrInf := PowerInfo{frames: make([]PowerInfoData, 10)}
	cmd := exec.Command("powerstat", "-R", "-c", "-z", "-f")
	data, err := cmd.Output()
	if err != nil {
		return err

	}
	output := string(data)
	lines := strings.Split(output, "\n")
	//var parsedLines [][]string
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

		} else if strings.HasPrefix(line, "C1") {

		} else if strings.HasPrefix(line, "C0") {

		} else if strings.HasPrefix("POLL", line) {

		} else {
			continue
		}
	}
	//parsedLines = append(parsedLines, strings.Split(line, " "))

	return nil

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
