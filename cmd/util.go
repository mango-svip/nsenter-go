package cmd

import (
    "context"
    "fmt"
    "github.com/docker/docker/api/types"
    dockerClient "github.com/docker/docker/client"
    "os"
    "os/exec"
    "strconv"
    "syscall"
)

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
