package cmd

import (
	"encoding/hex"
	"fmt"
	"github.com/LanceLRQ/zerotier-switcher/src/configs"
	"github.com/LanceLRQ/zerotier-switcher/src/tools"
	"log"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"
)

func CommandEntry(version string) {
	app := &cli.App{
		Name:     "zerotier-switcher",
		Usage:    "Zerotier Switcher",
		Commands: []*cli.Command{},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       configs.GetDefaultConfigPath(),
				Usage:       "Config file path",
				DefaultText: configs.GetDefaultConfigPath(),
			},
		},
		Action: func(c *cli.Context) error {
			if !(runtime.GOOS == "linux" || runtime.GOOS == "windows" || runtime.GOOS == "darwin") {
				return fmt.Errorf("unsupport operation system")
			}
			cfgFile := c.String("config")
			cfg, err := configs.ReadAppConfig(cfgFile)
			if err != nil {
				return err
			}
			planetFilePath := configs.GetPlanetFilePath(cfg)
			// 检查文件是否存在
			if _, err := os.Stat(planetFilePath); os.IsNotExist(err) {
				return fmt.Errorf("planet file (%s) not found", planetFilePath)
			}
			// 初始化操作
			if len(cfg.Planets) == 0 {
				// 如果配置文件里边没有数据，则先自动获取当前的planet信息并写入到文件中
				world, err := tools.ParsePlanetFile(planetFilePath)
				if err != nil {
					fmt.Println("init error: fail to read planet file.")
				} else {
					fmt.Printf("init: load planet file from %s\n", planetFilePath)
				}
				var root tools.Root
				var ep tools.InetAddress
				if len(world.Roots) > 0 {
					root = world.Roots[0]
					if len(root.StableEndpoints) > 0 {
						ep = root.StableEndpoints[0]
					}
				}
				cfg.Planets = append(cfg.Planets, configs.ZerotierPlanetFile{
					Hash:         hex.EncodeToString(world.Signature[:32]),
					Remark:       "Default",
					Data:         world.ToBase64String(),
					CreateTime:   world.Timestamp,
					WorldId:      world.ID,
					WorldType:    world.Type,
					RootIdentity: root.Identity.String(),
					RootEndpoint: ep.String(),
				})
				_ = configs.WriteAppConfig(cfgFile, cfg)
				fmt.Println("Press any key to continue")
				_, _ = fmt.Scanf("%s")
			}
			// TODO 交接给GUI

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
