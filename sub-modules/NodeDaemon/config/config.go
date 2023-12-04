package config

import (
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/utils/tools"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

/*
* ETCD_ADDR 由MasterNode获取，http接口传递给其他模块
* ETCD_PORT 由NodeDaemon向MasterNode配置，传递给其他模块，默认2379
* REDIS_ADDR 由MasterNode获取，http接口传递给其他模块
* REDIS_PORT 由NodeDaemon向MasterNode配置，http传递给其他模块，默认6379
* REDIS_PASSWORD 由NodeDaemon向MasterNode配置，http传递给其他模块，默认1145141919810
* REDIS_DB_INDEX 由NodeDaemon向MasterNode配置，http传递给其他模块，默认0
* DOCKER_HOST 由用户向NodeDaemon指派，传递给其他模块
 */

const (
	MasterNode  = "master"
	ServantNode = "servant"
)

const (
	RedisConfigurationUrl = "/platform/address/redis"
	EtcdConfigurationUrl  = "/platform/address/etcd"
)

var (
	StartMode            = MasterNode
	MasterAddress        = "127.0.0.1"
	InterfaceName        = "lo"
	EtcdAddr             = "127.0.0.1"
	EtcdPort             = 2379
	RedisAddr            = "127.0.0.1"
	RedisPort            = 6379
	RedisDBIndex         = 0
	RedisPassword        = "1145141919810"
	DockerHost           = "unix:///var/run/docker.sock"
	RedisImage           = "satellite_emulator/redis"
	EtcdImage            = "satellite_emulator/etcd"
	MasterNodeImage      = "satellite_emulator/master-node"
	InstanceManagerImage = "satellite_emulator/instance-manager"
	LinkManagerImage     = "satellite_emulator/link-manager"
)

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

func init() {
	DockerHost = GetConfigEnvString("DOCKER_HOST", DockerHost)
	StartMode = GetConfigEnvString("MODE", StartMode)
	InterfaceName = GetConfigEnvString("INTERFACE", InterfaceName)
	RedisImage = GetConfigEnvString("REDIS_IMAGE", RedisImage)
	EtcdImage = GetConfigEnvString("ETCD_IMAGE", EtcdImage)
	MasterNodeImage = GetConfigEnvString("MASTER_NODE_IMAGE", MasterNodeImage)
	InstanceManagerImage = GetConfigEnvString("INSTANCE_MANAGER_IMAGE", InstanceManagerImage)
	LinkManagerImage = GetConfigEnvString("LINK_MANAGER_IMAGE", LinkManagerImage)
	if StartMode == MasterNode {
		link, err := netlink.LinkByName(InterfaceName)
		if err != nil {
			errMsg := fmt.Sprintf("Unable to find Interface By Name %s: %s", InterfaceName, err.Error())
			logrus.Error(errMsg)
			panic(errMsg)
		}
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			errMsg := fmt.Sprintf("Unable to list addresses of %s: %s", InterfaceName, err.Error())
			logrus.Error(errMsg)
			panic(errMsg)
		}
		if len(addrs) <= 0 {
			errMsg := fmt.Sprintf("%s has no ipv4 address", InterfaceName)
			logrus.Error(errMsg)
			panic(errMsg)
		}
		MasterAddress = addrs[0].IP.String()
		EtcdAddr = MasterAddress
		RedisAddr = MasterAddress
	} else {
		MasterAddress = GetConfigEnvString("MASTER_ADDRESS", MasterAddress)
	}
}

func InitConfigMasterMode() error {

	EtcdPort = GetConfigEnvNumber("ETCD_PORT", EtcdPort)
	RedisPort = GetConfigEnvNumber("REDIS_PORT", RedisPort)
	RedisPassword = GetConfigEnvString("REDIS_PASSWORD", RedisPassword)
	RedisDBIndex = GetConfigEnvNumber("REDIS_DB_INDEX", RedisDBIndex)
	return nil
}

func InitConfigServantMode(masterAddr string) error {
	redisReqUrl := fmt.Sprintf("http://%s:8080%s", masterAddr, RedisConfigurationUrl)
	etcdReqUrl := fmt.Sprintf("http://%s:8080%s", masterAddr, EtcdConfigurationUrl)
	err := tools.DoWithRetry(func() error {
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

		RedisAddr = obj.Data.(map[string]interface{})["address"].(string)
		RedisPort = int(obj.Data.(map[string]interface{})["port"].(float64))
		RedisDBIndex = int(obj.Data.(map[string]interface{})["index"].(float64))
		RedisPassword = obj.Data.(map[string]interface{})["password"].(string)
		return nil
	}, 4)

	if err != nil {
		errMsg := fmt.Sprintf("Fetch Redis Configuration error: %s", err.Error())
		logrus.Error(errMsg)
		return err
	}

	err = tools.DoWithRetry(func() error {
		var obj ginmodel.JsonResp
		etcdResp, err := http.Get(etcdReqUrl)
		if err != nil {
			return err
		}
		err = json.NewDecoder(etcdResp.Body).Decode(&obj)
		if err != nil {
			return err
		}

		EtcdAddr = obj.Data.(map[string]interface{})["address"].(string)
		EtcdPort = int(obj.Data.(map[string]interface{})["port"].(float64))
		return nil
	}, 4)

	if err != nil {
		errMsg := fmt.Sprintf("Fetch Redis Configuration error: %s", err.Error())
		logrus.Error(errMsg)
		return err
	}
	return nil
}
