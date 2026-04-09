.PHONY: build clean

build:
	GOOS=windows GOARCH=amd64 go build -o FolderBackup.exe .

clean:
	rm -f FolderBackup.exe
