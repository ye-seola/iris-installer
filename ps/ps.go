package ps

import (
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type ProcessInfo struct {
	Name string
	PID  int
}

func GetProcessList() ([]ProcessInfo, error) {
	cmd := exec.Command("ps", "-eo", "name,pid")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe error: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("command start error: %w", err)
	}

	data, err := io.ReadAll(out)
	if err != nil {
		return nil, fmt.Errorf("read output error: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("command wait error: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var result []ProcessInfo

	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		result = append(result, ProcessInfo{Name: fields[0], PID: pid})
	}

	return result, nil
}
