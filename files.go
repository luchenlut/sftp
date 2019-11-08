package sftpClient

import (
	"fmt"
	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

// GetFile retreives a file from remote location
func GetFile(sftpClient *sftp.Client, localFilename string, remoteFilename string) error {
	stats, err := sftpClient.Lstat(remoteFilename)
	if err != nil {
		return fmt.Errorf("Could not determine type of file 'remote:%s': %v\n", remoteFilename, err)
	}
	if stats.IsDir() {
		if err := os.Mkdir(localFilename, 0755); err != nil {
			return fmt.Errorf("Could not create folder 'local:%s': %v\n", localFilename, err)
		}
		return nil
	}

	remote, err := sftpClient.Open(remoteFilename)
	if err != nil {
		return fmt.Errorf("Could not open file 'remote:%s': %v\n", remoteFilename, err)
	}
	defer remote.Close()

	local, err := os.Create(localFilename)
	if err != nil {
		return fmt.Errorf("Could not open file 'local:%s': %v\n", localFilename, err)
	}
	defer local.Close()

	if _, err := io.Copy(local, remote); err != nil {
		return fmt.Errorf("Could not copy 'remote:%s' to 'local:%s': %v\n", remoteFilename, localFilename, err)
	}
	log.Infof("download: %s", remoteFilename)

	return nil
}

// PutFile uploads a local file to remote location
func PutFile(sftpClient *sftp.Client, remoteFileName string, localFileName string) error {
	start := time.Now()
	localFile, err := os.Open(localFileName)
	if err != nil {
		return fmt.Errorf("Could not open file 'local:%s': %v\n", localFileName, err)
	}
	defer localFile.Close()

	stats, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("Could not stat file 'local:%s': %v\n", localFileName, err)
	}

	if stats.IsDir() {
		return CreateDirHierarchy(sftpClient, remoteFileName)
	}

	remoteFile, err := sftpClient.Create(remoteFileName)
	if err != nil {
		return fmt.Errorf("Could not create file 'remote:%s': %v\n", remoteFileName, err)
	}
	defer remoteFile.Close()

	if _, err := io.Copy(remoteFile, localFile); err != nil {
		fmt.Printf("Could not copy data from 'local:%s' to 'remote:%s': %v\n", localFileName, remoteFileName, err)
	}

	//if err := sftpClient.Chmod(remoteFileName, 0644); err != nil {
	//	return fmt.Errorf("Could not set file permissions on 'remote:%s': %v\n", remoteFileName, err)
	//}
	log.Infof("upload: %s (%v)", localFileName, time.Now().Sub(start))
	return nil
}

//copyToDest copies the src file to dst and delete local file
func copyToDest(src string, dst string) error {
	// Open the source file for reading
	s, err := os.Open(src)
	if err != nil {
		log.Printf("ERROR: copyToDest: Could not read file: %v\r\n", src)
		log.Printf("%v\r\n", err)
		return err
	}
	defer s.Close()

	// Open the destination file for writing
	d, err := os.Create(dst)
	if err != nil {
		log.Printf("ERROR: copyToDest: Could not create remote file: %v\r\n", dst)
		return err
	}

	// Copy the contents of the source file into the destination file
	if _, err := io.Copy(d, s); err != nil {
		log.Printf("ERROR: copyToDest: Error in copying file to %v\r\n", dst)
		d.Close()
		return err
	}

	// Return any errors that result from closing the destination file
	// Will return nil if no errors occurred
	if err := d.Close(); err != nil {
		log.Printf("ERROR: copyToDest: Error closing remote file: %v\r\n", dst)
	}
	if err := s.Close(); err != nil {
		log.Printf("ERROR: copyToDest: Error closing local file: %v\r\n", src)
	}

	log.Printf("copyToDest: File: %v copied to %v\r\n", src, dst)
	//return nil if success
	return nil
}

//moveToDest will use copyToDest to copy file and then delete the source file
func MoveToDest(src string, dst string) error {
	err := copyToDest(src, dst)
	if err != nil {
		log.Printf("ERROR: moveToDest: ERROR in copying file: %v\r\n", src)
		return err
	}
	log.Printf("moveToDest: Deleting File: %v\r\n", src)
	err1 := os.Remove(src)
	if err1 != nil {
		log.Printf("ERROR: moveToDest: Error in Deleting File: %v\r\n", src)
		return err1
	}
	return nil
}

//copyToDest copies the src file to dst and delete local file
func copyRemoteToDest(sftpClient *sftp.Client, src string, dst string) error {
	// Open the source file for reading
	s, err := sftpClient.Open(src)
	if err != nil {
		log.Printf("ERROR: copyToDest: Could not read file: %v\r\n", src)
		log.Printf("%v\r\n", err)
		return err
	}
	defer s.Close()

	// Open the destination file for writing
	d, err := sftpClient.Create(dst)
	if err != nil {
		log.Printf("ERROR: copyToDest: Could not create remote file: %v\r\n", dst)
		return err
	}

	// Copy the contents of the source file into the destination file
	if _, err := io.Copy(d, s); err != nil {
		fmt.Printf("Could not copy data from 'local:%s' to 'remote:%s': %v\n", s, d, err)
	}

	// Return any errors that result from closing the destination file
	// Will return nil if no errors occurred
	if err := d.Close(); err != nil {
		log.Printf("ERROR: copyToDest: Error closing remote file: %v\r\n", dst)
	}
	if err := s.Close(); err != nil {
		log.Printf("ERROR: copyToDest: Error closing local file: %v\r\n", src)
	}

	log.Printf("copyToDest: File: %v copied to %v\r\n", src, dst)
	//return nil if success
	return nil
}

//moveToDest will use copyToDest to copy file and then delete the source file
func MoveRemoteToDest(sftpClient *sftp.Client, src string, dst string) error {
	err := copyRemoteToDest(sftpClient, src, dst)
	if err != nil {
		log.Printf("ERROR: moveToDest: ERROR in copying file: %v\r\n", src)
		return err
	}
	log.Printf("moveToDest: Deleting File: %v\r\n", src)
	err1 := sftpClient.Remove(src)
	if err1 != nil {
		log.Printf("ERROR: moveToDest: Error in Deleting File: %v\r\n", src)
		return err1
	}
	return nil
}
