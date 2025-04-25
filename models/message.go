package models

import (
	"context"
	"encoding/json"
	"fmt"
	"ginchat/utils"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//chat 绑定Node和Conn 开启 sendProc和recvProc
//sendProc阻塞 recvProc node.Conn.ReadMessage()读取消息调用sendMsg或者sendGroupMsg
//sendMsg node, ok := clientMap[userId] node.DataQueue <- msg
//sendProc data <-node.DataQueue

// Message 消息
type Message struct {
	gorm.Model
	UserId     int64  //发送者
	TargetId   int64  //接受者
	Type       int    //发送类型  1私聊  2群聊  3心跳
	Media      int    //消息类型  1文字 2表情包 3语音 4图片 /表情包
	Content    string //消息内容
	CreateTime uint64 //创建时间
	ReadTime   uint64 //读取时间
	Pic        string
	Url        string
	Desc       string
	Amount     int //其他数字统计
}

func (message *Message) TableName() string {
	return "message"
}

func InitUDP() {
	//go udpSendProc()
	//go udpRecvProc()
	//fmt.Println("udp goroutine inited。。。。。。。。。。。。。。。")
}

//读写锁
var rwLocker sync.RWMutex

//var udpsendChan = make(chan []byte, 1024)

//	需要 :发送者ID ，接受者ID ，消息类型，发送的内容，发送类型
func Chat(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("chat::::::::::::::")
	//1.  获取参数 并 检验 token 等合法性
	//token := query.Get("token")
	query := request.URL.Query()
	Id := query.Get("userId")
	userId, _ := strconv.ParseInt(Id, 10, 64)
	//msgType := query.Get("type")
	//targetId := query.Get("targetId")
	//	context := query.Get("context")
	invalid := true //checkToke()  待.........
	conn, err := (&websocket.Upgrader{
		//token 校验
		CheckOrigin: func(r *http.Request) bool {
			return invalid
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	//2.获取conn
	currentTime := uint64(time.Now().Unix())
	node := &Node{
		Conn:          conn,
		Addr:          conn.RemoteAddr().String(), //客户端地址
		HeartbeatTime: currentTime,                //心跳时间
		LoginTime:     currentTime,                //登录时间
		DataQueue:     make(chan []byte, 50),
		GroupSets:     set.New(set.ThreadSafe),
	}
	//3. 用户关系
	//4. userid 跟 node绑定 并加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()
	//5.完成发送逻辑
	go sendProc(node)
	//6.完成接受逻辑
	go recvProc(node)
	//7.加入在线用户到redis缓存
	utils.Red.Set(context.Background(), "online_"+Id, []byte(node.Addr), time.Duration(viper.GetInt("timeout.RedisOnlineTime"))*time.Hour)
	//sendMsg(userId, []byte("欢迎进入聊天系统"))

}

//发送私聊消息
func sendMsg(userId int64, msg []byte) {
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()
	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	targetIdStr := strconv.Itoa(int(userId))
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime = uint64(time.Now().Unix())
	r, err := utils.Red.Get(ctx, "online_"+userIdStr).Result()
	if err != nil {
		fmt.Println(err)
	}
	if r != "" {
		if ok {
			node.DataQueue <- msg
			fmt.Println("sendMsg >>> userID: ", userId, "  msg:", string(msg))
		}
	}
	var key string
	if userId > jsonMsg.UserId {
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}
	res, err := utils.Red.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	score := float64(cap(res)) + 1
	ress, e := utils.Red.ZAdd(ctx, key, &redis.Z{score, msg}).Result() //jsonMsg
	//res, e := utils.Red.Do(ctx, "zadd", key, 1, jsonMsg).Result() //备用 后续拓展 记录完整msg
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println("影响行数:   ", ress)
}

//发送群聊消息
func sendGroupMsg(message Message, msg []byte) {
	fmt.Println("开始群发消息。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。。")
	userIds := SearchUserByGroupId(uint(message.TargetId))
	for i := 0; i < len(userIds); i++ {
		sendMsg(int64(userIds[i]), msg)
	}
}

func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			fmt.Println("[ws]sendProc >>>> msg :           ", string(data))
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println("WriteMessage err:", err)
				return
			}
		}
	}
}
func recvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println("ReadMessage err:", err)
			return
		}
		msg := Message{}
		msg.CreateTime = uint64(time.Now().Unix())
		err = json.Unmarshal(data, &msg)
		if err != nil {
			fmt.Println("Unmarshal err:", err)
		}
		//心跳检测 msg.Media == -1 || msg.Type == 3
		if msg.Type == 3 {
			currentTime := uint64(time.Now().Unix())
			node.Heartbeat(currentTime)
		} else {
			switch msg.Type {
			case 1:
				sendMsg(msg.TargetId, data) //聊
			case 2:
				sendGroupMsg(msg, data) //群发 发送的群ID ，消息内容
				// case 4: // 心跳
				//	node.Heartbeat()
			}
			//broadMsg(data) //todo 将消息广播到局域网
			fmt.Println("[ws] recvProc <<<<<               ", string(data))
		}
	}
}

//
//func broadMsg(data []byte) {
//	udpsendChan <- data
//}
//
////完成udp数据发送协程
//func udpSendProc() {
//	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
//		IP:   net.IPv4(192, 168, 0, 10),
//		Port: viper.GetInt("port.udp"),
//	})
//	if err != nil {
//		fmt.Println("DialUDP err=", err)
//	}
//	for {
//		select {
//		case data := <-udpsendChan:
//			fmt.Println("udpSendProc  data :", string(data))
//			_, err := con.Write(data)
//			if err != nil {
//				fmt.Println("data  write err=", err)
//				return
//			}
//		}
//	}
//	//defer con.Close()
//}
//
////完成udp数据接收协程
//func udpRecvProc() {
//	con, err := net.ListenUDP("udp", &net.UDPAddr{
//		IP:   net.IPv4zero,
//		Port: viper.GetInt("port.udp"),
//	})
//	if err != nil {
//		fmt.Println("ListenUDP err=", err)
//	}
//	for {
//		var buf [1024]byte
//		n, err := con.Read(buf[:])
//		if err != nil {
//			fmt.Println("data  read err=", err)
//			return
//		}
//		fmt.Println("udpRecvProc  data :", string(buf[0:n]))
//		dispatch(buf[0:n])
//	}
//	//defer con.Close()
//}

// RedisMsg 获取缓存里面的消息
func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) []string {
	rwLocker.RLock()
	//node, ok := clientMap[userIdA]
	rwLocker.RUnlock()
	//jsonMsg := Message{}
	//json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}
	//key = "msg_" + userIdStr + "_" + targetIdStr
	//rels, err := utils.Red.ZRevRange(ctx, key, 0, 10).Result()  //根据score倒叙

	var rels []string
	var err error
	if isRev {
		rels, err = utils.Red.ZRange(ctx, key, start, end).Result()
	} else {
		rels, err = utils.Red.ZRevRange(ctx, key, start, end).Result()
	}
	if err != nil {
		fmt.Println(err) //没有找到
	}
	// 后台通过websoket 推送消息
	/**
	for _, val := range rels {
		fmt.Println("sendMsg >>> userID: ", userIdA, "  msg:", val)
		node.DataQueue <- []byte(val)
	}**/
	return rels
}

// MarshalBinary 需要重写此方法才能完整的msg转byte[]
func (message *Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(message)
}
