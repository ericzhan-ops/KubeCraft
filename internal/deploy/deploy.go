package deploy

import (
	"fmt"
	"log"
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

	// 生成 Ansible inventory 文件到默认位置 /etc/ansible/hosts
	reporter.ReportProgress("生成 Ansible inventory 文件到默认位置...")
	err := config.GenerateDefaultInventory()
	if err != nil {
		return fmt.Errorf("failed to generate Ansible inventory: %v", err)
	}

	// 按顺序执行所有部署 playbook
	for i, playbook := range Playbooks {
		stepMsg := fmt.Sprintf("执行%s (%d/%d)...", playbook, i+1, len(Playbooks)+3)
		reporter.ReportProgress(stepMsg)

		err := executeAnsiblePlaybook(playbook)
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

	log.Println("Cluster deployment completed successfully")
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

// installAdditionalComponents 安装附加组件
func installAdditionalComponents(reporter ProgressReporter) error {
	reporter.ReportProgress("安装Helm...")
	log.Println("Installing Helm...")
	utils.InstallHelm()

	// 安装 Cilium
	reporter.ReportProgress("安装 Cilium...")
	log.Println("Installing Cilium...")
	err := installCilium()
	if err != nil {
		return fmt.Errorf("failed to install Cilium: %v", err)
	}

	// 安装 NFS CSI
	reporter.ReportProgress("安装 NFS CSI...")
	log.Println("Installing NFS CSI...")
	err = installNfsCsi()
	if err != nil {
		return fmt.Errorf("failed to install NFS CSI: %v", err)
	}

	// 安装 MetalLB
	reporter.ReportProgress("安装 MetalLB...")
	log.Println("Installing MetalLB...")
	err = installMetalLB()
	if err != nil {
		return fmt.Errorf("failed to install MetalLB: %v", err)
	}

	// 安装 Ingress Nginx
	reporter.ReportProgress("安装 Ingress Nginx...")
	log.Println("Installing Ingress Nginx...")
	err = installIngressNginx()
	if err != nil {
		return fmt.Errorf("failed to install Ingress Nginx: %v", err)
	}

	return nil
}

// installCilium 安装 Cilium 网络插件
func installCilium() error {
	// 这里应该使用 kubectl 应用 Cilium 的配置文件
	// 由于项目中没有提供 cilium 的 yaml 文件，我们暂时留空实现
	// 在实际应用中，这里会执行类似以下的命令：
	// kubectl apply -f https://raw.githubusercontent.com/cilium/cilium/main/install/kubernetes/quick-install.yaml

	// 示例实现:
	// cmd := exec.Command("kubectl", "apply", "-f", "https://raw.githubusercontent.com/cilium/cilium/main/install/kubernetes/quick-install.yaml")
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	//     return fmt.Errorf("failed to install Cilium: %v, output: %s", err, string(output))
	// }

	log.Println("Cilium installation placeholder - to be implemented")
	return nil
}

// installNfsCsi 安装 NFS CSI 驱动
func installNfsCsi() error {
	// 应用 NFS CSI 驱动相关的配置文件
	// 根据项目结构，yaml 文件位于 ./yaml 目录下
	csiFiles := []string{
		"../yaml/csi-nfs-rbac.yaml",
		"../yaml/csi-nfs-driverinfo.yaml",
		"../yaml/csi-nfs-controller.yaml",
		"../yaml/csi-nfs-node.yaml",
		"../yaml/storageclass-nfs.yaml",
	}

	for _, file := range csiFiles {
		cmd := exec.Command("kubectl", "apply", "-f", file)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to apply NFS CSI file %s: %v, output: %s", file, err, string(output))
		}
		log.Printf("Applied NFS CSI file: %s", file)
	}

	return nil
}

// installMetalLB 安装 MetalLB 负载均衡器
func installMetalLB() error {
	// 应用 MetalLB 配置文件
	cmd := exec.Command("kubectl", "apply", "-f", "./yaml/metallb.yaml")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply MetalLB: %v, output: %s", err, string(output))
	}

	log.Println("MetalLB applied successfully")
	return nil
}

// installIngressNginx 安装 Ingress Nginx 控制器
func installIngressNginx() error {
	// 应用 Ingress Nginx 配置文件
	cmd := exec.Command("kubectl", "apply", "-f", "./yaml/ingress-nginx.yaml")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply Ingress Nginx: %v, output: %s", err, string(output))
	}

	log.Println("Ingress Nginx applied successfully")
	return nil
}
