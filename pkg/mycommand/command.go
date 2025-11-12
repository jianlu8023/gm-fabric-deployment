package mycommand

import (
	"bytes"
	"io"
	"os/exec"
	"sync"
)

type OutPut struct {
	StdoutPut string
	StderrPut string
}

func Run(cmd *exec.Cmd) (OutPut, error) {

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return OutPut{}, err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return OutPut{}, err
	}

	if err = cmd.Start(); err != nil {
		return OutPut{}, err
	}
	// 获取两个管道的数据

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 将 stdoutPipe 的内容复制到 stdoutBuf
		// 忽略 io.Copy 的错误，因为主要的执行错误将由 cmd.Wait() 捕获
		_, _ = io.Copy(&stdoutBuf, stdoutPipe)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 将 stderrPipe 的内容复制到 stderrBuf
		// 忽略 io.Copy 的错误
		_, _ = io.Copy(&stderrBuf, stderrPipe)
	}()
	wg.Wait()

	// 等待执行结束
	err = cmd.Wait()

	// 获取两个管道里的数据
	var output OutPut
	output.StdoutPut = stdoutBuf.String()
	output.StderrPut = stderrBuf.String()

	return output, err
}
