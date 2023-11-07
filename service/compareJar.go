package service

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type RespData struct {
	NewJars    []string `json:"new_jars"`
	UpdateJars []string `json:"update_jars"`
	DeleteJars []string `json:"delete_jars"`
}

var keyPath = "./conf/private_key"

func listFilesAndMd5InDir(client *ssh.Client, dirPath string) (map[string]string, error) {
	session, err := client.NewSession()
	if err != nil {
		logger.SugarLog.Errorf("client NewSession failed,err:%v", err)
		return nil, err
	}

	defer session.Close()
	command := fmt.Sprintf(`find %s -type f -name "*.jar" -exec md5sum {} \;`, dirPath)
	output, err := session.CombinedOutput(command)
	if err != nil {
		logger.SugarLog.Errorf("session.CombinedOutput failed,err:%v", err)
		return nil, err
	}
	lines := string(output)

	fileMap := make(map[string]string)
	fileMd5Pairs := strings.Split(lines, "\n")
	for _, pair := range fileMd5Pairs {
		if pair == "" {
			continue
		}
		fields := strings.Fields(pair)
		if len(fields) >= 2 {
			md5Sum := fields[0]
			filePath := fields[1]

			//获取fileName
			fileName := filepath.Base(filePath)
			fileMap[fileName] = md5Sum
		}
	}

	return fileMap, nil
}

func CompareJar(p *model.ParamsCompareJar) (respData *RespData, err error) {

	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		logger.SugarLog.Errorf("ioutil.ReadFile failed,err:%v\n", err)
		return
	}

	singer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		logger.SugarLog.Errorf("ssh.ParsePrivateKey failed,err:%v\n", err)
		return
	}
	config := &ssh.ClientConfig{
		User: "tomcat",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(singer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	clientA, err := ssh.Dial("tcp", p.SrcHost+":22", config)
	if err != nil {
		logger.SugarLog.Errorf("ssh.Dial %s,err:%v\n", p.SrcHost, err)
		return
	}
	defer clientA.Close()

	clientB, err := ssh.Dial("tcp", p.DestHost+":22", config)
	if err != nil {
		logger.SugarLog.Errorf("ssh.Dial %s,err:%v\n", p.DestHost, err)
		return
	}
	defer clientB.Close()

	fileMapA, err := listFilesAndMd5InDir(clientA, p.SrcDir)
	if err != nil {
		logger.SugarLog.Errorf("listFilesAndMd5InDir failed,host:%s dir:%s,err:%v\n", p.SrcHost, p.SrcDir, err)
		return nil, fmt.Errorf("listFilesAndMd5InDir failed,host:%s dir:%s,err:%v", p.SrcHost, p.SrcDir, err)
	}

	fileMapB, err := listFilesAndMd5InDir(clientB, p.DestDir)
	if err != nil {
		logger.SugarLog.Errorf("listFilesAndMd5InDir failed,host:%s dir:%s,err:%v\n", p.DestDir, p.DestDir, err)
		return nil, fmt.Errorf("listFilesAndMd5InDir failed,host:%s dir:%s,err:%v", p.SrcHost, p.SrcDir, err)
	}

	//found new jar files
	var newJarNames []string

	for fileA, _ := range fileMapA {
		found := false
		for fileB, _ := range fileMapB {
			if fileA == fileB {
				found = true
				break
			}
		}

		if !found {
			newJarNames = append(newJarNames, fileA)
		}
	}

	//found delete and update jar files

	var deleteJarNames []string
	var updateJarNames []string

	for fileB, md5B := range fileMapB {
		found := false
		for fileA, md5A := range fileMapA {
			if fileB == fileA {
				found = true
				if md5A != md5B {
					updateJarNames = append(updateJarNames, fileB)
				}
				break
			}

		}

		if !found {
			deleteJarNames = append(deleteJarNames, fileB)
		}
	}

	respData = &RespData{
		NewJars:    newJarNames,
		UpdateJars: updateJarNames,
		DeleteJars: deleteJarNames,
	}
	return respData, nil

}
