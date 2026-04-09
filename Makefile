.PHONY: build clean

build:
	GOOS=windows GOARCH=amd64 go build -o FolderBackup.exe .
	GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui" -o FolderBackupSilent.exe .

clean:
	rm -f FolderBackup.exe FolderBackupSilent.exe
