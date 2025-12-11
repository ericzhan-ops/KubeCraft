package initialize

import (
	"KubeCraft/internal/utils"
	"fmt"
	"log"
	"os/exec"
	"time"
)

// ProgressReporter 进度报告接口
type ProgressReporter interface {
	ReportProgress(message string)
}

// Playbooks 定义初始化的 playbook 列表
var Playbooks = []string{
	"initConfigureHostname",
	"initUpdateEtcHosts",
	"initDisableServices",
	"initDisableSELinux",
	"initDisableSwap",
	"initUpdateSSHDConfig",
	"initUpdateLimitsConf",
	"initUpdateModulesConfig",
	"initConfigureTimeSync",
	"initConfigureSoftwareSources",
	"initInstallIPVS",
	"initConfigureKernel",
}

// Process 执行集群初始化过程
func Process(config utils.Config, reporter ProgressReporter) error {
	log.Println("Starting cluster initialization...")

	reporter.ReportProgress("检查Ansible安装状态...")
	// 安装 ansible（如果尚未安装）
	err := installAnsible(reporter)
	if err != nil {
		return fmt.Errorf("failed to install Ansible: %v", err)
	}

	// 生成 Ansible inventory 文件到默认位置 /etc/ansible/hosts
	reporter.ReportProgress("生成 Ansible inventory 文件到默认位置...")
	err = config.GenerateDefaultInventory()
	if err != nil {
		return fmt.Errorf("failed to generate Ansible inventory: %v", err)
	}

	// 按顺序执行所有初始化 playbook
	for i, playbook := range Playbooks {
		stepMsg := fmt.Sprintf("执行%s (%d/%d)...", playbook, i+1, len(Playbooks))
		reporter.ReportProgress(stepMsg)

		err := executeAnsiblePlaybook(playbook)
		if err != nil {
			return fmt.Errorf("failed to execute playbook %s: %v", playbook, err)
		}
	}

	reporter.ReportProgress("初始化完成")
	log.Println("Cluster initialization completed successfully")
	return nil
}

// installAnsible 安装 Ansible（如果尚未安装）
func installAnsible(reporter ProgressReporter) error {
	log.Println("Checking Ansible installation...")

	// 检查是否已安装 ansible
	if !isCommandAvailable("ansible") {
		reporter.ReportProgress("安装epel-release...")
		log.Println("Installing epel-release...")
		cmd1 := exec.Command("yum", "install", "epel-release", "-y")
		err := cmd1.Run()
		if err != nil {
			return fmt.Errorf("failed to install epel-release: %v", err)
		}

		reporter.ReportProgress("安装Ansible...")
		log.Println("Installing ansible...")
		cmd2 := exec.Command("yum", "install", "ansible", "-y")
		err = cmd2.Run()
		if err != nil {
			return fmt.Errorf("failed to install ansible: %v", err)
		}

		reporter.ReportProgress("复制Ansible配置文件...")
		log.Println("Copying ansible configuration...")
		err = utils.CopyFile("./ansiblePlaybook/ansible.cfg", "/etc/ansible/ansible.cfg")
		if err != nil {
			return fmt.Errorf("failed to copy ansible configuration: %v", err)
		}

		log.Println("Ansible installed successfully")
	} else {
		log.Println("Ansible is already installed")
	}

	return nil
}

// executeAnsiblePlaybook 执行指定的 Ansible Playbook
func executeAnsiblePlaybook(playbookName string) error {
	// 等待一段时间确保系统准备就绪
	time.Sleep(1 * time.Second)

	// 不再需要指定 -i 参数，因为使用默认的 /etc/ansible/hosts
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("cd ./ansiblePlaybook && ansible-playbook %s.yaml", playbookName))

	// 执行命令并返回结果
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}

	return nil
}

// isCommandAvailable 检查命令是否可用
func isCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
