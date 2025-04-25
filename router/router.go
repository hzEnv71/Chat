package router

import (
	"ginchat/docs"
	. "ginchat/service"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {

	r := gin.Default()
	//swagger
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	//静态资源
	r.Static("/asset", "asset/")
	r.StaticFile("/favicon.ico", "asset/images/favicon.ico")
	//	r.StaticFS()
	r.LoadHTMLGlob("views/**/*")

	//首页
	{
		r.GET("/", GetIndex)
		r.GET("/index", GetIndex)
		r.GET("/toRegister", ToRegister)
		r.GET("/toChat", ToChat)
		r.GET("/chat", Chat)
		r.POST("/searchFriends", SearchFriends)
	}

	//用户模块
	user := r.Group("/user")
	{
		user.POST("/getUserList", GetUserList)
		user.POST("/createUser", CreateUser)
		user.POST("/deleteUser", DeleteUser)
		user.POST("/updateUser", UpdateUser)
		user.POST("/findUserByNameAndPwd", FindUserByNameAndPwd)
		user.POST("/find", FindUserByID)
		//发送消息
		//user.GET("/sendMsg", SendMsg)
		//user.GET("/sendUserMsg", SendUserMsg)
		user.POST("/redisMsg", RedisMsg)
	}
	//上传文件
	r.POST("/upload", Upload)

	contact := r.Group("/contact")
	{ //添加好友
		contact.POST("/addFriend", AddFriend)
		//创建群
		contact.POST("/createCommunity", CreateCommunity)
		//群列表
		contact.POST("/loadCommunity", LoadCommunity)
		//加入群
		contact.POST("/joinGroup", JoinGroups)
	}
	//心跳续命 不合适  因为Node  所以前端发过来的消息再receProc里面处理
	//r.POST("/user/heartbeat", Heartbeat)

	return r
}
