package cmd

import (
    "context"
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    dockerClient "github.com/docker/docker/client"
    "strings"
)

var (
    listFlag          = false
    selectedContainer *types.Container
    keywordStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Background(lipgloss.Color("235"))
)

func init() {
    rootCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "list all containers")
}

type model struct {
    containers []types.Container
    cursor     int
}

func initialModel() model {
    ctx := context.Background()
    cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
    if err != nil {
        panic(err)
    }
    defer cli.Close()
    containers, err := cli.ContainerList(ctx, container.ListOptions{})

    tmp := make([]types.Container, 0)

    for _, c := range containers {
        if !strings.Contains(c.Names[0], "POD") {
            tmp = append(tmp, c)
        }
    }

    if err != nil {
        panic(err)
    }
    return model{
        containers: tmp,
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down":
            if m.cursor < len(m.containers)-1 {
                m.cursor++
            }
        case "enter":
            selectedContainer = &m.containers[m.cursor]
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m model) View() string {
    s := "选择要调试的容器 \n\n"
    for i, c := range m.containers {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }
        s += fmt.Sprintf("%s %s \n", cursor, keywordStyle.Render(c.Names[0][1:]))
    }

    s += "\n 按 q 退出。 \n"
    return s
}
