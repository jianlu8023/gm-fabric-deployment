package mycommand

import (
	"io"
	"os/exec"
)

type OutPut struct {
	StdoutPut string
	StderrPut string
}

func Run(cmd *exec.Cmd) (OutPut, error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return OutPut{}, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return OutPut{}, err
	}
	cmd.Start()
	// 获取两个管道的数据

	stdoutBytes, _ := io.ReadAll(stdoutPipe)
	stderrBytes, _ := io.ReadAll(stderrPipe)
	// 关闭管道
	stdoutPipe.Close()
	stderrPipe.Close()

	// 等待执行结束
	err = cmd.Wait()

	// 获取两个管道里的数据
	var output OutPut
	output.StdoutPut = string(stdoutBytes)
	output.StderrPut = string(stderrBytes)

	return output, err
}
