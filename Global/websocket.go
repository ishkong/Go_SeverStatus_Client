package Global

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"github.com/ishkong/go_ServerStatus_Client/Config"
)

func RunWebSocketClient(conf *Config.Config) {
	conn, _, err := websocket.DefaultDialer.Dial(conf.Server, nil)
	if err != nil {
		log.Warnf("连接到后端服务器 %v 发生错误：%v", conf.Server, err)
		time.Sleep(time.Millisecond * time.Duration(conf.Interval))
		go RunWebSocketClient(conf)
		return
	}
	handshake := fmt.Sprint(conf.Password)
	err = conn.WriteMessage(websocket.TextMessage, []byte(handshake))
	if err != nil {
		log.Warnf("反向WebSocket 握手时出现错误: %v", err)
	}
	_, retMsg, _ := conn.ReadMessage()
	if string(retMsg) != "Authentication success" {
		log.Warnf("鉴权失败，错误信息为: %s", retMsg)
		defer func(conn *websocket.Conn) {
			err := conn.Close()
			if err != nil {
				log.Warnf("关闭WebSocket时发生意外 %s", err)
			}
		}(conn)
		return
	}
	log.Info("连接成功")
	for {
		WsData := GenWebsocketMessage(conf.Interval)
		err = conn.WriteJSON(WsData)
		if err != nil {
			log.Warnf("上报消息发送错误：%v", err)
			go RunWebSocketClient(conf)
			return
		} else {
			log.Infof("上报消息成功 %v", WsData)
		}
		time.Sleep(time.Second * time.Duration(conf.Interval))
	}
}
