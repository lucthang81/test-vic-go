package sql

import (
	"bytes"
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/log"
	"io/ioutil"
	"os/exec"
	"strings"
)

func ReadSqlToDbIgnoreError(dataCenter *datacenter.DataCenter, path string) {
	content, err := ioutil.ReadFile(path)
	var removedCommentContent bytes.Buffer
	for _, lineString := range strings.Split(string(content), "\n") {
		if strings.Index(lineString, "--") != 0 {
			removedCommentContent.WriteString(lineString)
			removedCommentContent.WriteString("\n")
		}
	}
	queryCount := 0
	for _, queryString := range strings.Split(removedCommentContent.String(), ";\n") {
		queryCount++
		_, err = dataCenter.Db().Exec(queryString)
		if err != nil {
			log.Log("error %s at query %d. Query: %s", err.Error(), queryCount, queryString)
		}
	}
}

func ReadSqlToDb(dataCenter *datacenter.DataCenter, path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var removedCommentContent bytes.Buffer
	for _, lineString := range strings.Split(string(content), "\n") {
		if strings.Index(lineString, "--") != 0 {
			removedCommentContent.WriteString(lineString)
			removedCommentContent.WriteString("\n")
		}
	}

	queryCount := 0
	for _, queryString := range strings.Split(removedCommentContent.String(), ";\n") {
		queryCount++
		_, err = dataCenter.Db().Exec(queryString)
		if err != nil {
			log.Log("error %s at query %d. Query: %s", err.Error(), queryCount, queryString)
			return err
		}
	}
	return nil
}

func CreateDb(dbName string, owner string) (err error) {
	_, err = exec.Command("sh", "-c", fmt.Sprintf("createdb --username=gia --owner=%s --encoding=UTF8 %s;", owner, dbName)).Output()
	if err != nil {
		return err
	}
	return nil
}

func BackupTable(dbName string, tables []string) {
	tablesString := strings.Join(tables, " -t ")
	command := fmt.Sprintf("pg_dump -a -t %s %s", tablesString, dbName)

	// content, err := exec.Command("sh", "-c", command).Output()
	// if err != nil {
	// 	log.LogSerious("cannot backup %s %v", command, err)
	// 	return
	// }
	// _, filePath := log.CreateFileAndLog("backup.sql", string(content))
	// if filePath != "" {
	// 	log.LogWithAttach("Casino Backup", filePath)
	// }
	// return

	cmd := exec.Command("sh", "-c", command)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.LogSerious("cannot backup %s %v %v", command, err, stderr.String())
		return
	}
	_, filePath := log.CreateFileAndLog("backup.sql", out.String())
	if filePath != "" {
		log.LogWithAttach("Casino Backup", filePath)
	}
	return
}
