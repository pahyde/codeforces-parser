package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)


var rootCmd = &cobra.Command{
    Use: "forces",
    Short: "Forces is a CLI for parsing, testing, and submitting to codeforces",
    Long: `A codeforces parsing, testing, and submission CLI built for
           convenience and speed. Spend more time thinking and less on unnecessary keystrokes
           Read more at https://github.com/pahyde/forces`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("hello Cobra!")
    },
}


func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
