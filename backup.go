package sftpClient

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/pkg/sftp"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BackupFiles saves remote files to a local tar.gz file
func BackupFiles(sftpClient *sftp.Client, destination string, files []string) (saved <-chan bool, r <-chan ErrorResponse, done <-chan bool) {
	responseChannel := make(chan ErrorResponse)
	savedChannel := make(chan bool)
	doneChannel := make(chan bool)

	go func() {
		defer func() {
			doneChannel <- true
			close(responseChannel)
			close(savedChannel)
			close(doneChannel)
		}()

		var err error
		if _, err = os.Stat(destination); err != nil {
			if err2 := os.Mkdir(destination, 0755); err2 != nil {
				responseChannel <- ErrorResponse{Err: fmt.Errorf("Cannot access folder: %s; Failed to create folder: %s\n", err, err2)}
			}
		}

		tarFileName := fmt.Sprintf("%s.tar.gz", time.Now().UTC().Format("20060102-150405"))
		// create a file and get a handle to write gzipped data to
		tarPath := filepath.Join(destination, tarFileName)
		var zbuf *os.File
		if zbuf, err = os.Create(tarPath); err != nil {
			responseChannel <- ErrorResponse{err}
		}
		defer func() {
			if err := zbuf.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		// set up the gzip intermediary backing onto the file above
		gzw := gzip.NewWriter(zbuf)
		defer func() {
			if err := gzw.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		// get a handle to a tar writer instance using the gzip intermediate buffer above
		tw := tar.NewWriter(gzw)
		defer func() {
			if err := tw.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		for _, filename := range files {
			if err := TarFile(sftpClient, tw, filename); err != nil {
				responseChannel <- ErrorResponse{err}
			}
			savedChannel <- true
		}
	}()

	return savedChannel, responseChannel, doneChannel
}

// TarFile adds a remote file into a tar archive
func TarFile(sftpClient *sftp.Client, w *tar.Writer, filename string) error {
	f, err := sftpClient.Open(filename)
	if err != nil {
		return fmt.Errorf("Could not open file 'remote:%s': %v\n", filename, err)
	}
	defer f.Close()

	// create a tar header for this file
	stats, err := f.Stat()
	if err != nil {
		return fmt.Errorf("Could not stat file 'remote:%s': %v\n", filename, err)
	}
	hdr, err := tar.FileInfoHeader(stats, "")
	if err != nil {
		return fmt.Errorf("Could not create tar header for 'remote:%s': %v\n", filename, err)
	}
	name := strings.TrimPrefix(filename, "htdocs/sites/all")
	name = strings.TrimPrefix(name, "htdocs/sites/default")
	hdr.Name = name

	// write the header into our tar stream
	if err := w.WriteHeader(hdr); err != nil {
		return fmt.Errorf("Could not write tar header for 'remote:%s': %v\n", filename, err)
	}

	if !stats.IsDir() {
		// write the actual file to the tar stream
		if _, err := io.Copy(w, f); err != nil {
			return fmt.Errorf("Could not write tar data for 'remote:%s': %v\n", filename, err)
		}
	}

	return nil
}
