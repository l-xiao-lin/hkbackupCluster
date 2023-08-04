package service

import (
	"github.com/pkg/sftp"
	"hkbackupCluster/logger"
	"hkbackupCluster/pkg/sshclient"
	"io"
	"os"
	"path"
)

func UploadFile(localFilePath, remoteDir, remoteHost string) (err error) {
	conn, err := sshclient.SshConnect(remoteHost)
	if err != nil {
		return
	}
	defer conn.Close()

	//创建sftp会话
	client, err := sftp.NewClient(conn)
	if err != nil {
		return
	}
	defer client.Close()

	srcFile, err := os.Open(localFilePath)
	if err != nil {
		return
	}

	defer srcFile.Close()

	//获取需要上传的文件名称

	var remoteFileName = path.Base(localFilePath)

	//在服务器上创建文件并打开文件获得句柄
	destFile, err := client.Create(path.Join(remoteDir, remoteFileName))
	if err != nil {
		return
	}
	defer destFile.Close()

	buf := make([]byte, 1024)
	for {
		n, err := srcFile.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		destFile.Write(buf[:n])
	}
	logger.SugarLog.Infof("文件上传成功")
	return

}
