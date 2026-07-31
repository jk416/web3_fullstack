package global

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 是整个应用的配置总结构，对应 config/config.yaml 的顶层。
// 以后新增配置段（数据库、Redis、链节点...）就往这里加字段。
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

// JWTConfig 对应 yaml 里的 jwt 段。secret 生产应走环境变量 JWT_SECRET 覆盖。
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// ServerConfig 对应 yaml 里的 server 段。
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// DatabaseConfig 对应 yaml 里的 database 段，用来拼 Postgres 连接串(DSN)。
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// Conf 是全局唯一的配置实例：程序启动时由 LoadConfig 填充一次，之后各处只读。
// 类比 Spring：像一个被 @ConfigurationProperties 绑定好的单例 Bean，
// 区别是 Go 用包级变量直接访问（global.Conf.Server.Port），而不是靠 DI 容器 @Autowired 注入。
var Conf Config

func LoadConfig() error {
	viper.SetConfigFile("config/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	err = viper.Unmarshal(&Conf)
	if err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}
