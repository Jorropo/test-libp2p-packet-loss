package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	if runtime.GOOS != "linux" {
		panic("Only Linux is supported.")
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

	// Get tc's path
	tcPath, err := exec.LookPath("tc")
	if err != nil {
		panic(`could not find tc binary "tc": ` + err.Error())
	}

	tcTest := exec.Command(tcPath, "-V")
	tcTest.Stdin = null
	tcTest.Stdout = null
	tcTest.Stderr = null

	err = tcTest.Run()
	if err != nil {
		panic("could not run tc: " + err.Error())
	}

}
