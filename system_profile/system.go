package system_profile

import (
	"fmt"
	"github.com/vic/vic_go/log"
	"os"
	"runtime/pprof"
)

var folderPath string

func SetFolderPath(theFolderPath string) {
	folderPath = theFolderPath
}

func StartCPUProfile() {
	filePath := fmt.Sprintf("%s/casino_profile.prof", folderPath)
	f, err := os.Create(filePath)
	if err != nil {
		log.LogSerious("err start cpu profile %v", err)
	}
	pprof.StartCPUProfile(f)
}

func EndCPUProfile() {
	pprof.StopCPUProfile()
}

func OutputMemoryProfile() {

	filePath := fmt.Sprintf("%s/casino_profile.mprof", folderPath)
	f, err := os.Create(filePath)
	if err != nil {
		log.LogSerious("err output memory profile %v", err)
	}
	pprof.WriteHeapProfile(f)
	f.Close()
}
