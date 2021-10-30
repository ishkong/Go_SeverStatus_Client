package Config

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var currentPath = getCurrentPath()

var validate *validator.Validate

var DefaultConfigFile = path.Join(currentPath, "config.yml")

type Config struct {
	Server               string   `yaml:"server" validate:"required,uri"`
	User                 string   `yaml:"user" validate:"required"`
	Password             string   `yaml:"password" validate:"required"`
	Interval             uint16   `yaml:"interval" validate:"required,number,max=600,min=0"`
	InvalidInterfaceName []string `yaml:"invalid_interface_name" validate:"required"`
	LogAging             uint8    `yaml:"log_aging" validate:"required,number,max=255,min=0"`
	LogLevel             string   `yaml:"log_level" validate:"required"`
	LogForceNew          bool     `yaml:"log_force_new"`
}

func getCurrentPath() string {
	cwd, e := os.Getwd()
	if e != nil {
		panic(e)
	}
	return cwd
}

var (
	config *Config
	once   sync.Once
)

func Get() *Config {
	once.Do(func() {
		file, err := os.Open(DefaultConfigFile)
		config = &Config{}
		if err == nil {
			defer func() { _ = file.Close() }()
			if err = yaml.NewDecoder(file).Decode(config); err != nil {
				log.Fatal("配置文件错误", err)
			}
		} else {
			generateConfig()
			os.Exit(0)
		}
		validate = validator.New()
		errs := validate.Struct(config)
		if errs != nil {
			for _, err := range errs.(validator.ValidationErrors) {
				log.Fatalf("配置文件校验失败，配置 %s 的类型应为 %s ，错误为 %s", err.StructNamespace(), err.Tag(), err.Error())
			}
			os.Exit(0)
		}

	})
	return config
}

func generateConfig() {
	fmt.Println("未找到配置文件，正在为您生成配置文件中！")
	sb := strings.Builder{}
	sb.WriteString(defaultConfig)
	_ = os.WriteFile("config.yml", []byte(sb.String()), 0o644)
	fmt.Println("默认配置文件已生成，请修改 config.yml 后重新启动!")
	input := bufio.NewReader(os.Stdin)
	_, _ = input.ReadString('\n')
}

const defaultConfig = `server:
user:
password:
interval: 2
invalid_interface_name:
    - lo
    - tun
    - docker
    - kube
    - vmbr
    - br-
    - vnet
    - veth
log_aging: 15
log_level: info
log_force_new: false
`
