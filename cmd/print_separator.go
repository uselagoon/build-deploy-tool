package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var printCmd = &cobra.Command{
	Use:     "print",
	Aliases: []string{"p"},
	Short:   "Print out headers for sections like build steps, or warnings.",
}

var printBuildStepCompletion = &cobra.Command{
	Use:     "complete",
	Aliases: []string{"c"},
	Short:   "Will print out the header section for a build step completion",
	RunE: func(cmd *cobra.Command, args []string) error {
		buildStep, err := cmd.Flags().GetString("step")
		if err != nil {
			return fmt.Errorf("error reading step flag: %v", err)
		}
		startTime, err := cmd.Flags().GetString("start")
		if err != nil {
			return fmt.Errorf("error reading start flag: %v", err)
		}
		previousEnd, err := cmd.Flags().GetString("previous")
		if err != nil {
			return fmt.Errorf("error reading previous flag: %v", err)
		}
		tStartTime, err := time.Parse("2006-01-02 15:04:05 MST", startTime)
		if err != nil {
			return fmt.Errorf("error parsing start time: %v", err)
		}
		tPreviousEnd, err := time.Parse("2006-01-02 15:04:05 MST", previousEnd)
		if err != nil {
			return fmt.Errorf("error parsing previous time: %v", err)
		}
		now := time.Now()
		diff := time.Time{}.Add(time.Duration(now.Unix()-tPreviousEnd.Unix()) * time.Second)
		diffTotal := time.Time{}.Add(time.Duration(now.Unix()-tStartTime.Unix()) * time.Second)
		fmt.Println("##############################################")
		fmt.Println(fmt.Sprintf("STEP %s: Completed at %s Duration %s Elapsed %s", buildStep, now.Format("2006-01-02 15:04:05 MST"), diff.Format("15:04:05"), diffTotal.Format("15:04:05")))
		fmt.Println("##############################################")
		return nil
	},
}

var printBuildStepStart = &cobra.Command{
	Use:     "start",
	Aliases: []string{"s"},
	Short:   "Will print out the header section for a build step starting",
	RunE: func(cmd *cobra.Command, args []string) error {
		buildStep, err := cmd.Flags().GetString("step")
		if err != nil {
			return fmt.Errorf("error reading step flag: %v", err)
		}
		now := time.Now()
		fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++")
		fmt.Println(fmt.Sprintf("STEP %s: Started at %s", buildStep, now.Format("2006-01-02 15:04:05 MST")))
		fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++")
		return nil
	},
}

var printWarning = &cobra.Command{
	Use:     "warning",
	Aliases: []string{"w"},
	Short:   "Will print out the header section for a warning",
	RunE: func(cmd *cobra.Command, args []string) error {
		warningMsg, err := cmd.Flags().GetString("warning")
		if err != nil {
			return fmt.Errorf("error warning flag: %v", err)
		}
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println(fmt.Sprintf("WARNING: %s", warningMsg))
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		return nil
	},
}

var printBegin = &cobra.Command{
	Use:     "begin",
	Aliases: []string{"b"},
	Short:   "Will print out the beginning header",
	RunE: func(cmd *cobra.Command, args []string) error {
		step, err := cmd.Flags().GetString("step")
		if err != nil {
			return fmt.Errorf("error step flag: %v", err)
		}
		fmt.Println("##############################################")
		fmt.Println(fmt.Sprintf("BEGIN %s", step))
		fmt.Println("##############################################")
		return nil
	},
}

func init() {
	printCmd.AddCommand(printBuildStepCompletion)
	printCmd.AddCommand(printBuildStepStart)
	printCmd.AddCommand(printBegin)
	printCmd.AddCommand(printWarning)
	printBuildStepCompletion.Flags().StringP("step", "", "", "The step the build is currently at.")
	printBuildStepCompletion.Flags().StringP("start", "", "", "The time the build started.")
	printBuildStepCompletion.Flags().StringP("previous", "", "", "The time the previous step ended.")
	printBuildStepStart.Flags().StringP("step", "", "", "The step the build is currently at.")
	printBegin.Flags().StringP("step", "", "", "The step the build is currently at.")
	printWarning.Flags().StringP("warning", "", "", "The warning message to wrap.")
}
