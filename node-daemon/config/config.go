package config

import (
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/utils"
	"encoding/json"
	"fmt"
	"github.com/go-ini/ini"
	"net/http"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const (
	RedisConfigurationUrl = "/api/platform/address/redis"
	EtcdConfigurationUrl  = "/api/platform/address/etcd"
)

type AppConfigType struct {
	IsServant     bool   `ini:"IsServant"`
	MasterAddress string `ini:"MasterAddress"`
	InterfaceName string `ini:"InterfaceName"`
}

type DependencyConfigType struct {
	EtcdAddr       string `ini:"EtcdAddr"`
	EtcdPort       int    `ini:"EtcdPort"`
	RedisAddr      string `ini:"RedisAddr"`
	RedisPort      int    `ini:"RedisPort"`
	RedisDBIndex   int    `ini:"RedisDBIndex"`
	RedisPassword  string `ini:"RedisPassword"`
	DockerHostPath string `ini:"DockerHostPath"`
}

type GlobalConfigType struct {
	App        AppConfigType        `ini:"App"`
	Dependency DependencyConfigType `ini:"Dependency"`
}

var GlobalConfig GlobalConfigType

func GetConfigEnvNumber(name string, defaultVal int) int {
	env := os.Getenv(name)
	envNum, err := strconv.Atoi(env)
	if err != nil {
		return defaultVal
	}
	return envNum
}

func GetConfigEnvString(name string, defaultVal string) string {
	env := os.Getenv(name)
	if env == "" {
		return defaultVal
	}
	return env
}

func GetConfigEnvBool(name string, defaultVal bool) bool {
	env, err := strconv.ParseBool(os.Getenv(name))
	if err != nil {
		return defaultVal
	}
	return env
}

func InitConfig(iniPath string) {
	cfg, err := ini.Load(iniPath)
	if err != nil {
		errMsg := fmt.Sprintf("Load INI Config File in %s Error %s", iniPath, err.Error())
		logrus.Error(errMsg)
		panic(errMsg)
	}
	err = cfg.MapTo(&GlobalConfig)
	if err != nil {
		errMsg := fmt.Sprintf("Map INI Config File to Struct Error %s", err.Error())
		logrus.Error(errMsg)
		panic(errMsg)
	}
	GlobalConfig.Dependency.DockerHostPath = GetConfigEnvString("DOCKER_HOST", GlobalConfig.Dependency.DockerHostPath)
	GlobalConfig.App.IsServant = GetConfigEnvBool("IS_SERVANT", GlobalConfig.App.IsServant)
	GlobalConfig.App.InterfaceName = GetConfigEnvString("INTERFACE", GlobalConfig.App.InterfaceName)
	if !GlobalConfig.App.IsServant {
		link, err := netlink.LinkByName(GlobalConfig.App.InterfaceName)
		if err != nil {
			errMsg := fmt.Sprintf("Unable to find Interface By Name %s: %s", GlobalConfig.App.InterfaceName, err.Error())
			logrus.Error(errMsg)
			panic(errMsg)
		}
		addresses, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			errMsg := fmt.Sprintf("Unable to list addresses of %s: %s", GlobalConfig.App.InterfaceName, err.Error())
			logrus.Error(errMsg)
			panic(errMsg)
		}
		if len(addresses) <= 0 {
			errMsg := fmt.Sprintf("%s has no ipv4 address", GlobalConfig.App.InterfaceName)
			logrus.Error(errMsg)
			panic(errMsg)
		}
		GlobalConfig.App.MasterAddress = addresses[0].IP.String()
		GlobalConfig.Dependency.EtcdAddr = GlobalConfig.App.MasterAddress
		GlobalConfig.Dependency.RedisAddr = GlobalConfig.App.MasterAddress
	} else {
		GlobalConfig.App.MasterAddress = GetConfigEnvString("MASTER_ADDRESS", GlobalConfig.App.MasterAddress)
	}
}

func InitConfigMasterMode() error {

	GlobalConfig.Dependency.EtcdPort = GetConfigEnvNumber("ETCD_PORT", GlobalConfig.Dependency.EtcdPort)
	GlobalConfig.Dependency.RedisPort = GetConfigEnvNumber("REDIS_PORT", GlobalConfig.Dependency.RedisPort)
	GlobalConfig.Dependency.RedisPassword = GetConfigEnvString("REDIS_PASSWORD", GlobalConfig.Dependency.RedisPassword)
	GlobalConfig.Dependency.RedisDBIndex = GetConfigEnvNumber("REDIS_DB_INDEX", GlobalConfig.Dependency.RedisDBIndex)
	return nil
}

func InitConfigServantMode(masterAddr string) error {
	redisReqUrl := fmt.Sprintf("http://%s:8080%s", masterAddr, RedisConfigurationUrl)
	etcdReqUrl := fmt.Sprintf("http://%s:8080%s", masterAddr, EtcdConfigurationUrl)
	err := utils.DoWithRetry(func() error {
		var obj ginmodel.JsonResp
		redisResp, err := http.Get(redisReqUrl)
		if err != nil {
			return err
		}
		err = json.NewDecoder(redisResp.Body).Decode(&obj)

		if err != nil {
			logrus.Error("Unmashal Error")
			return err
		}

		GlobalConfig.Dependency.RedisAddr = obj.Data.(map[string]interface{})["address"].(string)
		GlobalConfig.Dependency.RedisPort = int(obj.Data.(map[string]interface{})["port"].(float64))
		GlobalConfig.Dependency.RedisDBIndex = int(obj.Data.(map[string]interface{})["index"].(float64))
		GlobalConfig.Dependency.RedisPassword = obj.Data.(map[string]interface{})["password"].(string)
		return nil
	}, 4)

	if err != nil {
		errMsg := fmt.Sprintf("Fetch Redis Configuration error: %s", err.Error())
		logrus.Error(errMsg)
		return err
	}

	err = utils.DoWithRetry(func() error {
		var obj ginmodel.JsonResp
		etcdResp, err := http.Get(etcdReqUrl)
		if err != nil {
			return err
		}
		err = json.NewDecoder(etcdResp.Body).Decode(&obj)
		if err != nil {
			return err
		}

		GlobalConfig.Dependency.EtcdAddr = obj.Data.(map[string]interface{})["address"].(string)
		GlobalConfig.Dependency.EtcdPort = int(obj.Data.(map[string]interface{})["port"].(float64))
		return nil
	}, 4)

	if err != nil {
		errMsg := fmt.Sprintf("Fetch Redis Configuration error: %s", err.Error())
		logrus.Error(errMsg)
		return err
	}
	return nil
}
