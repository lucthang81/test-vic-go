package test

//
//import (
//	"bytes"
//	"fmt"
//	"github.com/vic/vic_go/datacenter"
//	"github.com/vic/vic_go/log"
//	"io/ioutil"
//	"os"
//	"os/exec"
//	"path/filepath"
//	"strings"
//	// "log"
//)
//
//func CloneSchemaToTestDatabase(dbName string, sqlPaths []string) (err error) {
//	dataCenter := datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", "casino_x1_db", "")
//	dataCenter.Db().Exec(fmt.Sprintf("DROP DATABASE %s;", dbName))
//	dataCenter.Db().Close()
//
//	// _, err = exec.Command("sh", "-c", "createdb --username=gia --owner=avahardcore_user --encoding=UTF8 test_avahardcore_db; pg_dump avahardcore_db -s | psql test_avahardcore_db").Output()
//	_, err = exec.Command("sh", "-c", fmt.Sprintf("createdb --owner=vic_user --encoding=UTF8 %s;", dbName)).Output()
//	if err != nil {
//		fmt.Println("???")
//		fmt.Println(err)
//		return err
//	}
//	for _, sqlPath := range sqlPaths {
//		_, err = exec.Command("sh", "-c", fmt.Sprintf("psql -d %s -f %s", dbName, sqlPath)).Output()
//		if err != nil {
//			fmt.Println("??? ??")
//			fmt.Println(err)
//			return err
//		}
//	}
//	return nil
//}
//
//func CloneSchemaToTestDatabaseWithError(dbName string, sqlPaths []string) (err error) {
//	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(dir)
//
//	dataCenter := datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", "casino_x1_db", "")
//	dataCenter.Db().Exec(fmt.Sprintf("DROP DATABASE %s;", dbName))
//	dataCenter.Db().Close()
//
//	// _, err = exec.Command("sh", "-c", "createdb --username=gia --owner=avahardcore_user --encoding=UTF8 test_avahardcore_db; pg_dump avahardcore_db -s | psql test_avahardcore_db").Output()
//	_, err = exec.Command("sh", "-c", fmt.Sprintf("createdb --owner=vic_user --encoding=UTF8 %s;", dbName)).Output()
//	if err != nil {
//		fmt.Println("???")
//		fmt.Println(err)
//		return err
//	}
//	newDataCenter := datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", dbName, "")
//	defer newDataCenter.Db().Close()
//	for _, sqlPath := range sqlPaths {
//		err = ReadSqlToDb(newDataCenter, sqlPath)
//		if err != nil {
//			fmt.Println(err)
//			return err
//		}
//	}
//	return nil
//}
//
//func ReadSqlToDb(dataCenter *datacenter.DataCenter, path string) error {
//	content, err := ioutil.ReadFile(path)
//	if err != nil {
//		return err
//	}
//	var removedCommentContent bytes.Buffer
//	for _, lineString := range strings.Split(string(content), "\n") {
//		if strings.Index(lineString, "--") != 0 {
//			removedCommentContent.WriteString(lineString)
//			removedCommentContent.WriteString("\n")
//		}
//	}
//
//	queryCount := 0
//	for _, queryString := range strings.Split(removedCommentContent.String(), ";\n") {
//		queryCount++
//		_, err = dataCenter.Db().Exec(queryString)
//		if err != nil {
//			log.Log("error %s at query %d. Query: %s", err.Error(), queryCount, queryString)
//			return err
//		}
//	}
//	return nil
//}
//
//func CreateTestDatabase(dbName string) (err error) {
//	dataCenter := datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", "casino_x1_db", "")
//	dataCenter.Db().Exec(fmt.Sprintf("DROP DATABASE %s;", dbName))
//	dataCenter.Db().Close()
//
//	// _, err = exec.Command("sh", "-c", "createdb --username=gia --owner=avahardcore_user --encoding=UTF8 test_avahardcore_db; pg_dump avahardcore_db -s | psql test_avahardcore_db").Output()
//	_, err = exec.Command("sh", "-c", fmt.Sprintf("createdb --owner=vic_user --encoding=UTF8 %s;", dbName)).Output()
//	if err != nil {
//		fmt.Println(err)
//		return err
//	}
//	return nil
//}
//
//func DropTestDatabase(dbName string) {
//	dataCenter := datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", "casino_x1_db", "")
//	dataCenter.Db().Exec(fmt.Sprintf("DROP DATABASE %s;", dbName))
//	dataCenter.Db().Close()
//
//}
