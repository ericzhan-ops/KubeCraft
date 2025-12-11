package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Config 结构体定义配置参数
type Config struct {
	Masters             map[string]string `json:"masters"`
	Nodes               map[string]string `json:"nodes"`
	FirstMasterHostname string            `json:"firstMasterHostname"`
	SshUser             string            `json:"sshUser"`
	SshPass             string            `json:"sshPass"`
	SshPort             int               `json:"sshPort"`
	NetworkAdapter      string            `json:"networkAdapter"`
	KeepalivedVip       string            `json:"keepalivedVip"`
	ServiceNetwork      string            `json:"serviceNetwork"`
	PodNetwork          string            `json:"podNetwork"`
	LoadBalancerIP      string            `json:"loadBalancerIP"`
	NfsDir              string            `json:"nfsDir"`
	NfsServerIP         string            `json:"nfsServerIP"`
	OsType              string            `json:"osType"`
}

// GenerateDefaultInventory 生成 Ansible 默认的 inventory 文件并写入 /etc/ansible/hosts
func (config *Config) GenerateDefaultInventory() error {
	// 构造 INI 格式的 inventory 内容
	var builder strings.Builder

	// 添加注释说明
	builder.WriteString("# This is the default ansible 'hosts' file.\n")
	builder.WriteString("#\n")
	builder.WriteString("# It should live in /etc/ansible/hosts\n")
	builder.WriteString("#\n")
	builder.WriteString("#   - Comments begin with the '#' character\n")
	builder.WriteString("#   - Blank lines are ignored\n")
	builder.WriteString("#   - Groups of hosts are delimited by [header] elements\n")
	builder.WriteString("#   - You can enter hostnames or ip addresses\n")
	builder.WriteString("#   - A hostname/ip can be a member of multiple groups\n")
	builder.WriteString("\n")

	// 写入 masters 组
	builder.WriteString("# Kubernetes Master Nodes\n")
	builder.WriteString("[masters]\n")
	for hostname, ip := range config.Masters {
		if hostname != "" && ip != "" {
			builder.WriteString(fmt.Sprintf("%s ansible_host=%s\n", hostname, ip))
		}
	}

	// 写入 nodes 组
	builder.WriteString("\n# Kubernetes Worker Nodes\n")
	builder.WriteString("[nodes]\n")
	for hostname, ip := range config.Nodes {
		if hostname != "" && ip != "" {
			builder.WriteString(fmt.Sprintf("%s ansible_host=%s\n", hostname, ip))
		}
	}

	// 写入 all 组的变量
	builder.WriteString("\n# Global Variables\n")
	builder.WriteString("[all:vars]\n")
	builder.WriteString(fmt.Sprintf("ansible_ssh_port=%d\n", config.SshPort))
	builder.WriteString(fmt.Sprintf("ansible_ssh_user=%s\n", config.SshUser))
	builder.WriteString(fmt.Sprintf("ansible_ssh_pass=%s\n", config.SshPass))

	// 确保 /etc/ansible 目录存在
	err := os.MkdirAll("/etc/ansible", 0755)
	if err != nil {
		return fmt.Errorf("failed to create /etc/ansible directory: %v", err)
	}

	// 创建或截断 inventory 文件
	file, err := os.Create("/etc/ansible/hosts")
	if err != nil {
		return fmt.Errorf("failed to create inventory file: %v", err)
	}
	defer file.Close()

	// 写入 inventory 内容到文件
	_, err = file.WriteString(builder.String())
	if err != nil {
		return fmt.Errorf("failed to write inventory to file: %v", err)
	}

	return nil
}

// InstallHelm 安装 Helm
func InstallHelm() {
	if !IsCommandAvailable("helm") {
		log.Println("Helm not found, installing...")
		CopyFile("../pkg/helm", "/usr/local/bin/helm")
		log.Println("Helm installed successfully")
	} else {
		log.Println("Helm is already installed")
	}
}

// CopyFile 使用Go标准库复制文件
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}

// IsCommandAvailable 检查命令是否可用
func IsCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
