package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func main() {
	if runtime.GOOS != "linux" {
		panic("Only Linux is supported.")
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	null, err := os.Open("/dev/null")
	if err != nil {
		panic(err)
	}
	defer null.Close()

	// Get the command used to build go programs
	cmd, err := os.Open("/proc/self/cmdline")
	if err != nil {
		panic(err)
	}
	defer cmd.Close()

	goCmd, err := bufio.NewReader(cmd).ReadString(0)
	if err != nil {
		panic(err)
	}
	cmd.Close()
	if !strings.HasPrefix(goCmd, "go") {
		goCmd = "go"
	} else {
		goCmd = goCmd[:len(goCmd)-1] // trim null byte
	}

	goCmdPath, err := exec.LookPath(goCmd)
	if err != nil {
		panic(fmt.Sprintf("could not find go binary %q: %s", goCmd, err.Error()))
	}

	goTest := exec.Command(goCmdPath, "version")
	goTest.Stdin = null
	goTest.Stdout = null
	goTest.Stderr = null

	err = goTest.Run()
	if err != nil {
		panic("could not run go: " + err.Error())
	}

	// Builds
	fmt.Println("building client")
	build := exec.Command(goCmdPath, "build", ".")
	build.Dir = cwd + "/client"
	build.Stdin = null
	build.Stdout = os.Stdout
	build.Stderr = os.Stdout
	err = build.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println("building server")
	build = exec.Command(goCmdPath, "build", ".")
	build.Dir = cwd + "/server"
	build.Stdin = null
	build.Stdout = os.Stdout
	build.Stderr = os.Stdout
	err = build.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println("starting server")
	server := exec.Command(cwd + "/server/server")
	server.Stdin = null
	server.Stderr = os.Stderr
	out, err := server.StdoutPipe()
	if err != nil {
		panic(err)
	}
	defer out.Close()
	err = server.Start()
	if err != nil {
		panic(err)
	}
	defer server.Process.Kill()
	t := time.NewTimer(time.Second)
	c := make(chan struct{})
	go func() {
		select {
		case <-c:
		case <-t.C:
			out.Close()
		}
	}()
	got, err := io.ReadAll(out)
	close(c)
	t.Stop()
	out.Close()
	if err != nil {
		panic(err)
	}

	if string(got) != "running" {
		panic("didn't got expected \"running\" from server")
	}

	fmt.Println("server running\nstarting client")
	client := exec.Command(cwd+"/client/client", "/ip4/127.13.37.42/tcp/12345/p2p/12D3KooWNrc4Mm7jxnQ7FpraoDEZ3aAqF5QUzZwsGfgRRqw7asJG")
	client.Stdin = null
	client.Stdout = null
	client.Stderr = os.Stderr
	err = client.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println("success!")
}
