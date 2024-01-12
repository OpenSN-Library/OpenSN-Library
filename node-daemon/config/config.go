package config

import (
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/utils"
	"encoding/json"
	"fmt"
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
	IsServant        bool   `json:"is_servant"`
	MasterAddress    string `json:"master_address"`
	InterfaceName    string `json:"interface_name"`
	Debug            bool   `json:"debug"`
	InstanceCapacity int    `json:"instance_capacity"`
}

type DependencyConfigType struct {
	EtcdAddr       string `json:"etcd_addr"`
	EtcdPort       int    `json:"etcd_port"`
	RedisAddr      string `json:"redis_addr"`
	RedisPort      int    `json:"redis_port"`
	RedisDBIndex   int    `json:"redis_db_index"`
	RedisPassword  string `json:"redis_password"`
	DockerHostPath string `json:"docker_host_path"`
}

type GlobalConfigType struct {
	App        AppConfigType        `json:"app"`
	Dependency DependencyConfigType `json:"dependency"`
	Device     map[string][]string  `json:"device"`
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

func InitConfig(jsonPath string) {
	cfg, err := os.Open(jsonPath)
	if err != nil {
		errMsg := fmt.Sprintf("Load Json Config File in %s Error %s", jsonPath, err.Error())
		logrus.Error(errMsg)
		panic(errMsg)
	}
	decoder := json.NewDecoder(cfg)
	err = decoder.Decode(&GlobalConfig)
	if err != nil {
		errMsg := fmt.Sprintf("Map INI Config File to Struct Error %s", err.Error())
		logrus.Error(errMsg)
		panic(errMsg)
	}

	GlobalConfig.Dependency.DockerHostPath = GetConfigEnvString("DOCKER_HOST", GlobalConfig.Dependency.DockerHostPath)
	GlobalConfig.App.IsServant = GetConfigEnvBool("IS_SERVANT", GlobalConfig.App.IsServant)
	GlobalConfig.App.Debug = GetConfigEnvBool("DEBUG", GlobalConfig.App.Debug)
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
	logrus.Infof("Init Config Success, Config is %v", GlobalConfig)
	err = cfg.Close()
	if err != nil {
		errMsg := fmt.Sprintf("Close Config File Error: %s", err.Error())
		logrus.Error(errMsg)
		panic(errMsg)
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
