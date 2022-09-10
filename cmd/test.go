package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)


// forces test A
// forces test   <- tests most recently modified solution
var testCmd = &cobra.Command{
    Use: "test",
    Short: "",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("test.go")
    },
}

func init() {
    rootCmd.AddCommand(testCmd)
}

