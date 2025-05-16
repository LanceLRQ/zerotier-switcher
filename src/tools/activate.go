package tools

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/LanceLRQ/zerotier-switcher/src/configs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ReplacePlanetAndJoinNetwork 替换 planet 文件并加入指定网络
func ReplacePlanetAndJoinNetwork(base64Planet string, networkID string, callback func(int, string)) error {
	// 1. 解码 base64 planet 数据
	callback(1, "Decoding planet")
	planetData, err := base64.StdEncoding.DecodeString(base64Planet)
	if err != nil {
		return fmt.Errorf("base64 decode error: %v", err)
	}

	// 2. 获取 planet 文件路径
	callback(2, "Get planet path")
	planetPathFolder, err := configs.GetZerotierProfileFolder()
	if err != nil {
		return fmt.Errorf("get planet file path error: %v", err)
	}

	planetPath := path.Join(planetPathFolder, "planet")

	// 3. 检查是否已是当前planet
	callback(3, "Checking planet file")
	newHash := md5.Sum(planetData)
	newHashStr := hex.EncodeToString(newHash[:])
	existingHashStr, err := getFileHash(planetPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("check planet file error: %v", err)
	}
	if existingHashStr == newHashStr {
		return fmt.Errorf("same planet file, abort")
	}

	// 4. 写入新的 planet 文件
	callback(4, "Writing planet file")
	if err := os.WriteFile(planetPath, planetData, 0644); err != nil {
		return fmt.Errorf("write planet file error: %v", err)
	}

	// 5. 重启 ZeroTier 服务
	callback(5, "Restarting zerotier service, please wait")
	if err := restartZeroTierService(); err != nil {
		return fmt.Errorf("restart zerotier service error: %v", err)
	}

	if networkID != "" {
		callback(6, "Restarting zerotier service, please wait")
		// 6. 加入指定网络
		if err := joinZeroTierNetwork(networkID); err != nil {
			return fmt.Errorf("join network error: %v", err)
		}
		callback(7, "Done")
	} else {
		callback(7, "Done")
	}

	return nil
}

// getFileHash 计算文件的 MD5 哈希
func getFileHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:]), nil
}

// restartZeroTierService 重启 ZeroTier 服务
func restartZeroTierService() error {
	switch runtime.GOOS {
	case "linux":
		// 尝试 systemd
		if _, err := exec.LookPath("systemctl"); err == nil {
			if err := exec.Command("systemctl", "restart", "zerotier-one").Run(); err == nil {
				return nil
			}
		}
		// 尝试 service 命令
		if _, err := exec.LookPath("service"); err == nil {
			if err := exec.Command("service", "zerotier-one", "restart").Run(); err == nil {
				return nil
			}
		}
		// 尝试直接 kill 和启动
		exec.Command("pkill", "zerotier-one").Run()
		return exec.Command("zerotier-one", "-d").Start()

	case "darwin":
		// macOS
		exec.Command("launchctl", "unload", "/Library/LaunchDaemons/com.zerotier.one.plist").Run()
		exec.Command("launchctl", "load", "/Library/LaunchDaemons/com.zerotier.one.plist").Run()
		return nil

	case "windows":
		// Windows
		exec.Command("net", "stop", "ZeroTier One").Run()
		time.Sleep(2 * time.Second)
		return exec.Command("net", "start", "ZeroTier One").Run()

	default:
		return fmt.Errorf("unsupport operation system: %s", runtime.GOOS)
	}
}

// joinZeroTierNetwork 加入 ZeroTier 网络
func joinZeroTierNetwork(networkID string) error {
	// 清理网络ID，移除可能的前后空格和非字母数字字符
	cleanID := strings.TrimSpace(networkID)
	if len(cleanID) != 16 {
		return fmt.Errorf("network id error")
	}

	// 等待服务完全启动
	time.Sleep(3 * time.Second)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command(filepath.Join(os.Getenv("ProgramFiles"), "ZeroTier", "One", "zerotier-cli.bat"), "join", cleanID)
	default:
		cmd = exec.Command("zerotier-cli", "join", cleanID)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: %v, stderr: %s", err, string(output))
	}

	return nil
}

func GetCurrentPlanetHashFromOS() string {
	planetPathFolder, err := configs.GetZerotierProfileFolder()
	if err != nil {
		return ""
	}

	planetPath := path.Join(planetPathFolder, "planet")
	existingHashStr, err := getFileHash(planetPath)
	if err != nil && !os.IsNotExist(err) {
		return ""
	}
	return existingHashStr
}

func CheckIsCurrentPlanet(base64Planet, existingHashStr string) bool {
	planetData, err := base64.StdEncoding.DecodeString(base64Planet)
	if err != nil {
		return false
	}
	newHash := md5.Sum(planetData)
	newHashStr := hex.EncodeToString(newHash[:])
	return existingHashStr != "" && newHashStr != "" && existingHashStr == newHashStr
}
