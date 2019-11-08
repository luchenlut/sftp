package main

import (
	"fmt"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"sftp-client"
	"strings"
	"time"
)

const (
	HOSTNAME = "192.168.1.184"
	PORT     = 22
	USERNAME = "sftptest"
	PASSWORD = "12345678"

	//本地文件路径
	LOCALPATH = "D:\\var\\log\\collect\\"
	//远程文件路径
	REMOTEPATH = "/home/luchen/data/"
	//处理文件中转路径
	PROCESSINGPATH = "D:\\var\\log\\processing\\"
	//执行完成文件路径
	DONEPATH = "D:\\var\\log\\done\\"
)

func main() {
	start := time.Now()
	SC, err := sftpClient.New(USERNAME, PASSWORD, HOSTNAME, PORT)
	if err != nil {
		logrus.Error(err)
	}

	//fmt.Println(SC)

	uploadFile2Remote(SC) //1.上传文件

	//downloadRemoteFile(SC)	//2.下载文件

	//for key, value := range findRemoteFile(SC)  {
	//	fmt.Println(key,value)
	//	//sftpClient.MoveRemoteToDest(SC,value,value+".bin")
	//	sftpClient.MoveRemoteToDest(SC,value,strings.Replace(value,".tmp",".bin",-1))
	//}					//3.发现远程文件
	//for key, value := range findLocalFile()  {
	//	fmt.Println(key,value)
	//}					//4.发现本地文件

	//moveFile()		//5.移动文件
	fmt.Println(time.Now().Sub(start))
}

func uploadFile2Remote(sc *sftp.Client) {
	localFiles, err := sftpClient.FindAllLocalFiles([]string{LOCALPATH})
	if err != nil {
		logrus.Error(err)
	}
	for key, value := range localFiles {
		println(key, value)
		if strings.Contains(value, ".zip") || strings.Contains(value, ".txt") {
			//将文件移动到上传目录
			tmpName := value + ".tmp"
			err = sftpClient.MoveToDest(LOCALPATH+value, PROCESSINGPATH+tmpName)
			if err != nil {
				logrus.Error("ERROR in moving file to Done folder")
			}
			//上传文件
			upload(sc, tmpName)
			//上传成功文件去掉.tmp
			err = sftpClient.MoveRemoteToDest(sc, REMOTEPATH+tmpName, REMOTEPATH+strings.Replace(tmpName, ".tmp", ".bin", -1))
			if err != nil {
				logrus.Error("ERROR in moving remote file to Done folder")
			}
			//移除上传成功文件
			err = sftpClient.MoveToDest(PROCESSINGPATH+tmpName, DONEPATH+tmpName)
			if err != nil {
				logrus.Error("ERROR in moving file to Done folder")
			}
		}
	}
}

func upload(sc *sftp.Client, value string) {
	err := sftpClient.PutFile(sc, REMOTEPATH+value, PROCESSINGPATH+value)
	if err != nil {
		logrus.Error(err)
	}
}

func downloadRemoteFile(SC *sftp.Client) {
	remoteFiles := findRemoteFile(SC)
	for key, value := range remoteFiles {
		println(key, value)
		if strings.Contains(value, ".json") || strings.Contains(value, ".txt") {
			//将文件移动到下载目录
			/*			err = sftpClient.MoveToDest(LOCALPATH+value, PROCESSINGPATH+value)
						if err != nil {
							logrus.Error("ERROR in moving file to Done folder")
						}*/
			//下载文件
			err := sftpClient.GetFile(SC, LOCALPATH+value[18:], value)
			if err != nil {
				logrus.Error(err)
			}
			//移除下载成功文件
			//err = sftpClient.MoveToDest(PROCESSINGPATH+value, DONEPATH+value)
			//if err != nil {
			//	logrus.Error("ERROR in moving file to Done folder")
			//}
		}
	}
}

func findRemoteFile(SC *sftp.Client) []string {
	//文件发现
	remoteFiles, err := sftpClient.FindAllRemoteFiles(SC, []string{REMOTEPATH})
	if err != nil {
		logrus.Error(err)
	}

	files := make([]string, 0)
	for _, value := range remoteFiles {
		files = append(files, value)
	}
	return files
}

func findLocalFile() []string {
	//文件发现
	localFiles, err := sftpClient.FindAllLocalFiles([]string{LOCALPATH})
	if err != nil {
		logrus.Error(err)
	}

	files := make([]string, 0)
	for _, value := range localFiles {
		files = append(files, value)
	}
	return files
}
