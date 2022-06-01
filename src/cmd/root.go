/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
	"sync"

	"github.com/arejula27/measurepymemo/pkg/docker"
	"github.com/arejula27/measurepymemo/pkg/frecuenzy"
	"github.com/arejula27/measurepymemo/pkg/http"
	"github.com/arejula27/measurepymemo/pkg/powerstat"
	"github.com/spf13/cobra"
)

const commandName = "measurepymemo"

var (
	rootFlags struct {
		file      string
		paralel   int
		frecuenzy int
		count     int
		message   string
		maxTime   int
		test      bool
		url       bool
		threshold int
		governor  string
	}

	stopMeasurement chan bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   commandName,
	Short: "measure  energy",
	Long:  `Measure the energy and other CPU stats via software while running a container`,

	Run: measurepymemo,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	flags := rootCmd.Flags()
	flags.StringVarP(&rootFlags.file,
		"file",
		"f",
		"",
		"Write the wide kind of metrics on the specified file")

	flags.IntVarP(&rootFlags.paralel,
		"paralel",
		"p",
		1,
		"Set how many paralel executions of the container")
	flags.IntVarP(&rootFlags.frecuenzy, //TODO
		"frecuenzy",
		"F",
		0,
		"Set the frecuenzy of the CPU for measuring")
	flags.IntVarP(&rootFlags.count,
		"count",
		"c",
		1,
		"Set the number of secuencial executions the cointainer will have. When -p is used all paralel executions will do -c iterations")
	flags.StringVarP(&rootFlags.message,
		"message",
		"m",
		"",
		"Append a message on the output")

	flags.IntVarP(&rootFlags.maxTime,
		"time",
		"t",
		60,
		"Set the maximun time in seconds the program will gather metrics, if the container lates more the output will not be correct. Ensure the max time is correctly set")

	flags.BoolVar(
		&rootFlags.url,
		"remote",
		false,
		"Set the http request it would call instead of running a container")

	flags.IntVar(&rootFlags.threshold, //TODO
		"threshold",
		127,
		"Set the threshold")

	flags.StringVarP(&rootFlags.governor,
		"governor",
		"g",
		"",
		"Set the governor")

	//TODO flag for choose image (i)
}

func measurepymemo(cmd *cobra.Command, args []string) {

	if rootFlags.governor != "" {
		fm := frecuenzy.New()
		err := fm.SetGovernor(rootFlags.governor)
		if err != nil {
			fmt.Println("Problems on set a new governor: " + rootFlags.governor)
			log.Fatalln(err)

		}
		defer func() {
			err = fm.Restore()
			if err != nil {
				fmt.Println("Problems recoveryn previous governor")

			}

		}()

	} else if rootFlags.frecuenzy > 0 {
		fm := frecuenzy.New()
		err := fm.SetFrecuenzy(rootFlags.frecuenzy)
		if err != nil {
			fmt.Println("Problems on set a new frecuenzy")
			panic(err)
		}
		defer func() {
			err = fm.Restore()
			if err != nil {
				fmt.Println("Problems recoveryn previous governor")
			}

		}()
	}

	wg := new(sync.WaitGroup)
	
	stopMeasurement = make(chan bool, 1)
	checkPrivileges()

	if rootFlags.url {
		wg.Add(1)
		go callHttp(wg)
	} else {
		wg.Add(1)
		go launchContainer(wg)
	}

	wg.Wait()

}

func gatherMetrics(mainWg *sync.WaitGroup) {

	measurer := powerstat.New(strconv.Itoa(rootFlags.maxTime))

	go func() {
		<-stopMeasurement
		measurer.End()
	}()

	pwrInf, err := measurer.Run()
	pwrInf.Message = rootFlags.message
	if err != nil {
		fmt.Println(commandName + " could not gather any metric")
		os.Exit(1)
	}
	fmt.Print(pwrInf.GetHeader())
	fmt.Print(pwrInf.GetData())
	if rootFlags.file != "" {
		err = WriteFile(rootFlags.file, pwrInf)
		if err != nil {
			fmt.Println("Error writing the output on the file")

		}

	}
	mainWg.Done()
	//fmt.Println("Error, mediciones terminadas antes del proceso, ponga más tiempo de medición")
	//os.Exit(1)
}

func launchContainer(mainWg *sync.WaitGroup) {
	wg := new(sync.WaitGroup)

	for i := 0; i < rootFlags.paralel; i++ {
		wg.Add(1)
		go func(id int, wg *sync.WaitGroup) {
			for j := 0; j < rootFlags.count; j++ {
				err := docker.RunContainer("arejula27/pymemo:test")
				if err != nil {
					fmt.Println("Error al lanzar el contenedor")
					os.Exit(1)
				}
				fmt.Printf(" contenedor %d secuencia %d finalizado\n", j, id)
			}

			wg.Done()
		}(i, wg)
	}

	wg.Wait()
	stopMeasurement <- true
	mainWg.Done()

}

func callHttp(mainWg *sync.WaitGroup) {
	wg := new(sync.WaitGroup)
	for i := 0; i < rootFlags.paralel; i++ {
		wg.Add(1)
		go func(id int, wg *sync.WaitGroup) {
			for j := 0; j < rootFlags.count; j++ {

				err := http.CallRemote(strconv.Itoa(rootFlags.threshold))
				if err != nil {
					fmt.Println("Error al lanzar realizar la llamada http")
					os.Exit(1)
				}
				fmt.Printf(" llamada http %d secuencia %d finalizado\n", j, id)
			}

			wg.Done()
		}(i, wg)
	}

	wg.Wait()
	stopMeasurement <- true
	mainWg.Done()

}

func checkPrivileges() {
	currentUser, _ := user.Current()
	if currentUser.Username != "root" {

		fmt.Println("It is not posible gather metrcis without root privileges")
		os.Exit(1)

	}
}

func WriteFile(fileName string, metrics powerstat.PowerInfo) error {
	var newfile bool
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		newfile = true
		file, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	}
	if err != nil {
		return err
	}
	defer file.Close()
	if newfile {
		_, err = file.WriteString(metrics.GetCsvHeader())
		if err != nil {
			return err
		}
	}
	_, err = file.WriteString(metrics.ToCsv())
	if err != nil {
		return err
	}

	return nil
}
