package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var CliRoot = &cobra.Command{
	Use: "fondness-app",
	Short: "An fondness blockchain application",
	Long: `....`,
}

func Execute() {
	if err := CliRoot.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

