package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

type ZerotierSwitcherProfile struct {
	Planets             []ZerotierPlanetFile `json:"planets"`
	ZerotierProfilePath string               `json:"zerotier_profile_path"` // custom zerotier profile path
}

type ZerotierPlanetFile struct {
	Hash         string `json:"hash"`
	Remark       string `json:"remark"` // remark text (view)
	Data         string `json:"data"`   // base64 encoded text
	CreateTime   uint64 `json:"create_time"`
	WorldId      uint64 `json:"world_id"`
	WorldType    uint8  `json:"world_type"` // (1=Planet, 127=Moon)
	RootIdentity string `json:"root_identity"`
	RootEndpoint string `json:"root_endpoint"` // Ip address of the planet file (view)

	AutoJoinNetwork string `json:"auto_join_network"`
}

// GetDefaultConfigPath 获取当前程序的配置文件默认路径
func GetDefaultConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	appPath := fmt.Sprintf("%s/zerotier-switcher", dir)
	if err := os.MkdirAll(appPath, 0o755); err != nil {
		return ""
	}
	return path.Join(appPath, "profile.json")
}

// GetPlanetFilePath 获取Zerotier的planet文件位置
func GetPlanetFilePath(cfg *ZerotierSwitcherProfile) string {
	return fmt.Sprintf("%s/planet", cfg.ZerotierProfilePath)
}

func GetZerotierProfileFolder() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return "/var/lib/zerotier-one", nil
	case "darwin":
		return "/Library/Application Support/ZeroTier/One", nil
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "ZeroTier", "One"), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func GetDefaultZerotierSwitcherProfile() ZerotierSwitcherProfile {
	profileFolder, _ := GetZerotierProfileFolder()
	return ZerotierSwitcherProfile{
		Planets:             []ZerotierPlanetFile{},
		ZerotierProfilePath: profileFolder,
	}
}

// ReadAppConfig 读取配置
func ReadAppConfig(path string) (*ZerotierSwitcherProfile, error) {
	data, err := os.ReadFile(path)
	cfg := GetDefaultZerotierSwitcherProfile()
	if os.IsNotExist(err) {
		err = WriteAppConfig(path, &cfg)
		return &cfg, err
	} else if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &cfg)
	return &cfg, err
}

// WriteAppConfig 写入配置
func WriteAppConfig(path string, config *ZerotierSwitcherProfile) error {
	// 获取配置文件路径
	data, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
