package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var codeCmd = &cobra.Command{
    Use: "code",
    Short: "",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("hello from Cobra!")
    },
}

func init() {
    rootCmd.AddCommand(codeCmd)
}

