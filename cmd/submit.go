package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
    Use: "submit",
    Short: "",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("hello from Cobra!")
    },
}

func init() {
    rootCmd.AddCommand(submitCmd)
}

