package main

import (
	"ginchat/models"
	"ginchat/router"
	"ginchat/utils"
	"github.com/spf13/viper"
	"time"
)

func main() {
	utils.InitConfig()
	utils.InitMySQL()
	utils.InitRedis()

	//初始化定时器
	delay := time.Duration(viper.GetInt("timeout.DelayHeartbeat")) * time.Second
	tick := time.Duration(viper.GetInt("timeout.HeartbeatHz")) * time.Second
	utils.Timer(delay, tick, models.CleanConnection, "")

	models.InitUDP() //初始化udp携程
	r := router.Router()
	addr := viper.GetString("port.server")
	r.Run(addr)

}
