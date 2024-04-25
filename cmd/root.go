package cmd

import (
    "errors"
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/spf13/cobra"
    "os"
)

var rootCmd = &cobra.Command{
    Use:   "nsenter-go",
    Short: "attach container network space",
    RunE: func(cmd *cobra.Command, args []string) error {

        if listFlag {
            p := tea.NewProgram(initialModel())
            if _, err := p.Run(); err != nil {
                fmt.Printf("Alas, there's been an error: %v", err)
                os.Exit(1)
            }
            if selectedContainer != nil {
                enterNsenter(getPidByContainer(*selectedContainer))
            }
            return nil
        }
        if containerName != "" {
            enterNsenter(getPidByContainer(filterContainer(containerName)))
            return nil
        }

        return errors.New("unrecognized command")
    },
}

func Execute() {
    rootCmd.Execute()
}
