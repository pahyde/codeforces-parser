package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
    Use: "test",
    Short: "",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("hello from Cobra!")
    },
}

func init() {
    rootCmd.AddCommand(testCmd)
}

