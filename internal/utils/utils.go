package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
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

// Inventory 结构体定义 Ansible inventory 格式
type Inventory struct {
	Masters struct {
		Hosts interface{} `json:"hosts"`
	} `json:"masters"`
	Nodes struct {
		Hosts interface{} `json:"hosts"`
	} `json:"nodes"`
	All struct {
		Hosts    interface{} `json:"hosts"`
		Children []string    `json:"children"`
		Vars     struct {
			AnsibleSSHPort int    `json:"ansible_ssh_port"`
			AnsibleSSHUser string `json:"ansible_ssh_user"`
			AnsibleSSHPass string `json:"ansible_ssh_pass"`
		} `json:"vars"`
	} `json:"all"`
}

// GenerateInventory 生成 Ansible inventory 并写入临时文件
func (config *Config) GenerateInventory() (string, error) {
	// 提取 master 和 node 的 IPs
	var masterIPs []string
	for _, ip := range config.Masters {
		masterIPs = append(masterIPs, ip)
	}

	var nodeIPs []string
	for _, ip := range config.Nodes {
		nodeIPs = append(nodeIPs, ip)
	}

	// 构造 inventory 对象
	inventory := &Inventory{
		Masters: struct {
			Hosts interface{} `json:"hosts"`
		}{
			Hosts: masterIPs,
		},
		Nodes: struct {
			Hosts interface{} `json:"hosts"`
		}{
			Hosts: nodeIPs,
		},
		All: struct {
			Hosts    interface{} `json:"hosts"`
			Children []string    `json:"children"`
			Vars     struct {
				AnsibleSSHPort int    `json:"ansible_ssh_port"`
				AnsibleSSHUser string `json:"ansible_ssh_user"`
				AnsibleSSHPass string `json:"ansible_ssh_pass"`
			} `json:"vars"`
		}{
			Hosts:    []interface{}{},
			Children: []string{"masters", "nodes"},
			Vars: struct {
				AnsibleSSHPort int    `json:"ansible_ssh_port"`
				AnsibleSSHUser string `json:"ansible_ssh_user"`
				AnsibleSSHPass string `json:"ansible_ssh_pass"`
			}{
				AnsibleSSHPort: config.SshPort,
				AnsibleSSHUser: config.SshUser,
				AnsibleSSHPass: config.SshPass,
			},
		},
	}

	// 将 inventory 转换为 JSON
	inventoryData, err := json.Marshal(inventory)
	if err != nil {
		return "", fmt.Errorf("failed to marshal inventory: %v", err)
	}

	// 创建临时文件存储 inventory
	tempFile, err := os.CreateTemp("", "inventory-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	// 写入 inventory 数据到临时文件
	_, err = tempFile.Write(inventoryData)
	if err != nil {
		return "", fmt.Errorf("failed to write inventory to temp file: %v", err)
	}

	return tempFile.Name(), nil
}

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
