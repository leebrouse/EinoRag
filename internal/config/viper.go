// 配置模块：集中初始化 Viper，支持 YAML 文件与环境变量覆盖
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// 在进程启动时加载配置。若读取失败直接 panic，避免在不确定配置下继续运行。
func init() {
	if err := NewViperConfig(); err != nil {
		panic(err)
	}
}

var once sync.Once

// 单例初始化：确保配置只初始化一次，避免多次读取或竞态
func NewViperConfig() (err error) {
	once.Do(func() {
		err = newViperConfig()
	})
	return
}

func newViperConfig() error {
	// 计算相对路径：以调用方工作目录为基准，定位到本文件所在目录，用于查找 global.yaml
	relPath, err := getRelativePathFromCaller()
	if err != nil {
		return err
	}
	// 配置文件设置：名称、类型、查找路径
	viper.SetConfigName("global")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(relPath)
	// 环境变量设置：将环境变量下划线转为配置键中的连字符；开启自动读取环境变量
	viper.EnvKeyReplacer(strings.NewReplacer("_", "-"))
	viper.AutomaticEnv()
	// 保留示例：如需 Stripe 相关 ENV 绑定，可按需启用
	// _ = viper.BindEnv("stripe-key", "STRIPE_KEY", "endpoint-stripe-secret", "ENDPOINT_STRIPE_SECRET")
	// 绑定环境变量到配置键 gemini.apikey，支持常见变量名
	_ = viper.BindEnv("gemini.apikey", "GEMINI_API_KEY", "GOOGLE_API_KEY", "GOOGLE_KEY")
	// 读取 YAML 配置文件（若同名环境变量存在，会覆盖文件值）
	return viper.ReadInConfig()
}

// 根据调用者工作目录与当前文件路径，计算相对目录。用于向 Viper 添加配置文件查找路径。
func getRelativePathFromCaller() (relPath string, err error) {
	callerPwd, err := os.Getwd()
	if err != nil {
		return
	}
	_, here, _, _ := runtime.Caller(0)
	relPath, err = filepath.Rel(callerPwd, filepath.Dir(here))
	// 调试输出：定位配置读取路径，可按需移除
	fmt.Printf("caller from: %s, here: %s, relpath: %s", callerPwd, here, relPath)
	return
}
