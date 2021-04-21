package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func GetHomeDir() string {
	usr, _ := user.Current()
	return usr.HomeDir
}

func NormalizePath(inputPath string) string {
	if inputPath == "~" {
		return GetHomeDir()
	} else if strings.HasPrefix(inputPath, "~/") {
		return filepath.Join(GetHomeDir(), inputPath[2:])
	}

	return inputPath
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func Int32Ptr(i int32) *int32 { return &i }

func Prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
