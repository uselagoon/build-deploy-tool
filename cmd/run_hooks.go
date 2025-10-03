package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/hooks"
)

var runCmd = &cobra.Command{
	Use:     "run",
	Aliases: []string{"r"},
	Short:   "Run commands",
}

var runHookCmd = &cobra.Command{
	Use:     "hooks",
	Aliases: []string{"hook", "h"},
	Short:   "Run build hooks",
	Long:    "Run build hooks of specific types at entrypoints in a build",
	Run: func(cmd *cobra.Command, args []string) {
		hookName, err := cmd.Flags().GetString("hook-name")
		if err != nil {
			fmt.Printf("error reading hook-name flag: %v\n", err)
			os.Exit(1)
		}
		hookDir, err := cmd.Flags().GetString("hook-directory")
		if err != nil {
			fmt.Printf("error reading hook-directory flag: %v\n", err)
			os.Exit(1)
		}
		dir := fmt.Sprintf("/kubectl-build-deploy/hooks/%s", hookDir)
		err = hooks.RunHooks(hookName, dir)
		if err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	runCmd.AddCommand(runHookCmd)
	runHookCmd.Flags().StringP("hook-name", "N", "", "The name of the hooks run")
	runHookCmd.Flags().StringP("hook-directory", "D", "", "The name of the hook directory to run")
}
