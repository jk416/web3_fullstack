package global

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 是整个应用的配置总结构，对应 config/config.yaml 的顶层。
// 以后新增配置段（数据库、Redis、链节点...）就往这里加字段。
type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

// ServerConfig 对应 yaml 里的 server 段。
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// Conf 是全局唯一的配置实例：程序启动时由 LoadConfig 填充一次，之后各处只读。
// 类比 Spring：像一个被 @ConfigurationProperties 绑定好的单例 Bean，
// 区别是 Go 用包级变量直接访问（global.Conf.Server.Port），而不是靠 DI 容器 @Autowired 注入。
var Conf Config

// TODO(you): 在这里（或本包内新文件）写 LoadConfig，用 viper 读 config/config.yaml 填充 Conf。

func LoadConfig() error {
	viper.SetConfigFile("config/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	err = viper.Unmarshal(&Conf)
	if err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}
