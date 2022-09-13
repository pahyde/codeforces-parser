package cmd

import (
    "fmt"
    "os"
    "path/filepath"
    "log"
    "github.com/spf13/cobra"
)

// forces test A
// forces test   <- tests most recently modified solution
var testCmd = &cobra.Command{
    Use: "test",
    Short: "",
    Args: cobra.MaximumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        configDir, err := os.UserConfigDir()
        if err != nil {
            log.Fatal(err)
        }
        appDir := filepath.Join(configDir, "forces")

        // read session.json data
        p := filepath.Join(appDir, "session.json")
        var session Session
        if err := readJSON(p, &session); err != nil {
            log.Fatal(err)
        }

        // read templates.json data
        p = filepath.Join(appDir, "templates.json")
        var registry TemplateRegistry
        if err := readJSON(p, &registry); err != nil {
            log.Fatal(err)
        }

        fmt.Println(session)
        fmt.Println(registry)
    },
}

func init() {
    rootCmd.AddCommand(testCmd)
}

