package cmd

import (
    "github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
    Use: "test",
    Short: "",
//    Run: func(cmd *cobra.Command, args []string) {
//        fmt.Println("test.go")
//    },
}

func init() {
    rootCmd.AddCommand(testCmd)
}

