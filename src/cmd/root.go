/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"
	"os"
	"time"

	"github.com/arejula27/measurepymemo/pkg/powerstat"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "measurepymemo",
	Short: "mide energía",
	Long:  `mide la energía usada por un programa`,

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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.measurepymemo.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func measurepymemo(cmd *cobra.Command, args []string) {
	//err := docker.RunContainer("arejula27/pymemo:test")
	measurer := powerstat.New("60")
	go func() {
		time.Sleep(time.Second * 10)
		measurer.End()
	}()

	pwrInf, err := measurer.Run()
	if err != nil {
		panic(err)
	}

	err = WriteFile("data.csv", pwrInf.ToCsv())
	if err != nil {
		log.Fatalln(err)

	}

}

func printHeader() string {
	header := "Average power(Watts);Average frecuenzy;"
	header += "Max power(Watts);Max frecuenzy;"
	header += "Min power(Watts);Min frecuenzy;"
	header += "C2 resident;C2 count;C2 latency;"
	header += "C1 resident;C1 count;C1 latency;"
	header += "C0 resident;C0 count;C0 latency;"
	header += "POLL resident;POLL count;POLL latency"
	header += "\n"
	return header
}

func WriteFile(fileName, output string) error {
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
		_, err = file.WriteString(printHeader())
		if err != nil {
			return err
		}
	}
	_, err = file.WriteString(output)
	if err != nil {
		return err
	}

	return nil
}
