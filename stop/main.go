package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	stopped := false

	cmd1 := exec.Command("taskkill", "/IM", "FolderBackupSilent.exe", "/F")
	if output, err := cmd1.CombinedOutput(); err == nil {
		fmt.Println("FolderBackupSilent.exe stopped.")
		stopped = true
	} else {
		fmt.Printf("FolderBackupSilent.exe: %s", string(output))
	}

	cmd2 := exec.Command("taskkill", "/IM", "FolderBackup.exe", "/F")
	if output, err := cmd2.CombinedOutput(); err == nil {
		fmt.Println("FolderBackup.exe stopped.")
		stopped = true
	} else {
		fmt.Printf("FolderBackup.exe: %s", string(output))
	}

	if !stopped {
		fmt.Println("\nNo backup process was running.")
	}

	fmt.Println("\nPress Enter to close...")
	buf := make([]byte, 1)
	os.Stdin.Read(buf)
}
