package main

import (
	"fmt"
	"os/exec"
)

func main() {
	cmd := exec.Command("taskkill", "/IM", "FolderBackupSilent.exe", "/F")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Could not stop process: %s\n", string(output))
	} else {
		fmt.Println("FolderBackupSilent.exe stopped.")
	}
}
