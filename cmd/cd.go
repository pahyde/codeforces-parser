package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var cdCmd = &cobra.Command{
    Use: "cd",
    Short: "",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("hello from Cobra!")
    },
}

func init() {
    rootCmd.AddCommand(cdCmd)
}

