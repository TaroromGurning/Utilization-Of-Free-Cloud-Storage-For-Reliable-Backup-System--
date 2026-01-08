package service

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type RcloneResult struct {
	Success          bool
	Output           string
	ErrorMsg         string
	Duration         time.Duration
	TransferredBytes int64
}

func ExecuteCliJob(commandArgs []string) RcloneResult {
	// 1. Safety Check: Pastikan command tidak kosong agar tidak panic
	if len(commandArgs) == 0 {
		return RcloneResult{
			Success:  false,
			ErrorMsg: "Command arguments cannot be empty",
		}
	}

	startTime := time.Now()

	cmdName := commandArgs[0]
	args := commandArgs[1:]

	cmd := exec.Command(cmdName, args...)

	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	outputStr := strings.TrimSpace(string(output))

	result := RcloneResult{
		Duration: duration,
		Output:   outputStr,
	}

	// Parse bytes baik sukses maupun gagal (kadang rclone error tapi sempat transfer data)
	result.TransferredBytes = parseTransferredBytes(outputStr)

	if err != nil {
		result.Success = false
		result.ErrorMsg = fmt.Sprintf("Exit Error: %v. Output: %s", err, result.Output)
		return result
	}

	result.Success = true
	return result
}
