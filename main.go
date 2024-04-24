package main

import (
    "context"
    "flag"
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/filters"
    dockerClient "github.com/docker/docker/client"
    "os"
    "os/exec"
    "strconv"
    "strings"
    "syscall"
)

var (
    listFlag          = false
    containerName     = ""
    selectedContainer *types.Container
    keywordStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Background(lipgloss.Color("235"))
)

func init() {
    flag.StringVar(&containerName, "c", "", "-c <container_name>")
    flag.BoolFunc("l", "list container", func(s string) error {
        listFlag = true
        return nil
    })
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

func main() {
    flag.Parse()
    if containerName == "" && !listFlag {
        flag.Usage()
        os.Exit(-1)
    }

    if listFlag {
        p := tea.NewProgram(initialModel())
        if _, err := p.Run(); err != nil {
            fmt.Printf("Alas, there's been an error: %v", err)
            os.Exit(1)
        }
        if selectedContainer != nil {
            enterNsenter(getPidByContainer(*selectedContainer))
        }
        return
    }

    pid := getPidByContainer(filterContainer())
    enterNsenter(pid)
}

func filterContainer() types.Container {
    // list container
    cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
    if err != nil {
        panic(err)
    }
    defer cli.Close()
    containers, err := cli.ContainerList(context.Background(), container.ListOptions{Filters: filters.NewArgs(filters.Arg("name", containerName))})
    if err != nil {
        panic(err)
    }
    if len(containers) == 0 {
        panic("未找到指定容器: " + containerName)
    }
    return containers[0]
}

func getPidByContainer(c types.Container) int {
    cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
    if err != nil {
        panic(err)
    }
    defer cli.Close()

    inspectInfo, err := cli.ContainerInspect(context.Background(), c.ID)
    if err != nil {
        panic(err)
    }
    fmt.Printf("获取到容器[%s] pid: [%s] \n", keywordStyle.Render(c.Names[0][1:]), keywordStyle.Render(strconv.Itoa(inspectInfo.State.Pid)))
    return inspectInfo.State.Pid
}

func enterNsenter(pid int) {
    // 获取当前用户的 UID 和 GID
    uid := uint32(os.Getuid())
    gid := uint32(os.Getgid())

    // 创建新进程的属性
    attr := &syscall.SysProcAttr{
        Credential: &syscall.Credential{
            Uid: uid,
            Gid: gid,
        },
    }

    cmd := exec.Command("nsenter", "-n", "-t", strconv.Itoa(pid))

    cmd.SysProcAttr = attr
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err := cmd.Start()
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    err = cmd.Wait()
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
}
