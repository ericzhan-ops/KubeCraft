package utils

import (
	"log"
	"os/exec"
)

// InstallHelm 安装 Helm
func InstallHelm() {
	if !IsCommandAvailable("helm") {
		log.Println("Helm not found, installing...")
		exec.Command("\\cp", "/usr/local/ymctl/pkg/helm", "/usr/local/bin/").Run()
		log.Println("Helm installed successfully")
	} else {
		log.Println("Helm is already installed")
	}
}

// IsCommandAvailable 检查命令是否可用
func IsCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
