package deploy

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"KubeCraft/internal/utils"
)

// ProgressReporter 进度报告接口
type ProgressReporter interface {
	ReportProgress(message string)
}

// Playbooks 定义部署的 playbook 列表
var Playbooks = []string{
	"installContainerd",
	"installNginx",
	"installKeepalived",
	"installKubeInit",
	"installKubeJoin",
	"installKubePost",
}

// Process 执行集群部署过程
func Process(config utils.Config, reporter ProgressReporter) error {
	log.Println("Starting cluster deployment...")

	// 生成 Ansible inventory 文件
	reporter.ReportProgress("生成 Ansible inventory 文件...")
	inventoryFile, err := config.GenerateInventory()
	if err != nil {
		return fmt.Errorf("failed to generate Ansible inventory: %v", err)
	}
	// 确保在函数结束时删除临时文件
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			return
		}
	}(inventoryFile)

	// 按顺序执行所有部署 playbook
	for i, playbook := range Playbooks {
		stepMsg := fmt.Sprintf("执行%s (%d/%d)...", playbook, i+1, len(Playbooks))
		reporter.ReportProgress(stepMsg)

		err := executeAnsiblePlaybook(playbook, inventoryFile)
		if err != nil {
			return fmt.Errorf("failed to execute playbook %s: %v", playbook, err)
		}
	}

	// 安装附加组件
	reporter.ReportProgress("安装附加组件...")
	log.Println("Installing additional components...")
	err = installAdditionalComponents(reporter)
	if err != nil {
		return err
	}

	reporter.ReportProgress("部署完成")
	log.Println("Cluster deployment completed successfully")
	return nil
}

// executeAnsiblePlaybook 执行指定的 Ansible Playbook
func executeAnsiblePlaybook(playbookName string, inventoryFile string) error {
	// 等待一段时间确保系统准备就绪
	time.Sleep(1 * time.Second)

	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("cd /usr/local/ymctl/ansiblePlaybook && ansible-playbook -i %s %s.yaml", inventoryFile, playbookName))

	// 执行命令并返回结果
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}

	return nil
}

// installAdditionalComponents 安装附加组件
func installAdditionalComponents(reporter ProgressReporter) error {
	reporter.ReportProgress("安装Helm...")
	log.Println("Installing Helm...")
	utils.InstallHelm()

	// 可以在这里添加其他组件的安装
	// installCilium()
	// installNfsCsi()
	// installMetalLB()
	// installIngressNginx()

	return nil
}
