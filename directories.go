package sftpClient

import (
	"fmt"
	"github.com/pkg/sftp"
	"os"
	"strings"
)

// CreateDir creates a folder on remote location
func CreateDir(sftpClient *sftp.Client, remoteFolderPath string) error {
	if _, err := sftpClient.Lstat(remoteFolderPath); err != nil {
		if os.IsNotExist(err) {
			if err := sftpClient.Mkdir(remoteFolderPath); err != nil {
				return fmt.Errorf("Could not create folder 'remote:%s': %v\n", remoteFolderPath, err)
			}
			if err := sftpClient.Chmod(remoteFolderPath, 0755); err != nil {
				return fmt.Errorf("Could not set folder permissions on 'remote:%s': %v\n", remoteFolderPath, err)
			}
		} else {
			return fmt.Errorf("Error finding 'remote:%s': %v\n", remoteFolderPath, err)
		}
	}
	return nil
}

// CreateDirHierarchy creates a folder hierarchy on remote location
func CreateDirHierarchy(sftpClient *sftp.Client, remoteFolderPath string) error {
	parent := "."
	if strings.HasPrefix(remoteFolderPath, "/") {
		parent = "/"
		remoteFolderPath = strings.TrimPrefix(remoteFolderPath, "/")
	}

	tree := strings.Split(remoteFolderPath, "/")

	for _, dir := range tree {
		parent = strings.Join([]string{parent, dir}, "/")
		if err := CreateDir(sftpClient, parent); err != nil {
			return err
		}
	}
	return nil
}
