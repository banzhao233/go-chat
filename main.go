package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type user struct {
	name  string
	pwd   string
	token string
}
type userMessage struct {
	Username string `json:"username"`
	Message  string `json:"message"`
	Token    string `json:"token"`
}

var users []*user = []*user{
	{
		name: "yxlr",
		pwd:  "114514",
	},
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var history string

const d = "{\n    \"users\": {\n        \"用户名\":\"密码\",\n        \"用户名2\" : \"密码2\"\n    }\n}"
const h = "<!DOCTYPE html>\n<html>\n\n<head>\n    <meta charset=\"UTF-8\">\n    <title>聊天界面</title>\n    <style>\n        * {\n            box-sizing: border-box;\n        }\n\n        body {\n            margin: 0;\n            font-family: 'Montserrat', sans-serif;\n            background-color: #f7f7f7;\n        }\n\n        .overlay {\n            position: fixed;\n            top: 0;\n            left: 0;\n            width: 100%;\n            height: 100%;\n            background-color: rgba(0, 0, 0, 0.5);\n            display: flex;\n            justify-content: center;\n            align-items: center;\n            z-index: 9999;\n        }\n\n        .modal {\n            background-color: #fff;\n            padding: 20px;\n            border-radius: 10px;\n            box-shadow: 0 0 20px rgba(0, 0, 0, 0.2);\n            display: flex;\n            flex-direction: column;\n            align-items: center;\n            max-width: 500px;\n            width: 100%;\n        }\n\n        .modal h3 {\n            font-size: 24px;\n            font-weight: 600;\n            margin: 0 0 20px;\n        }\n\n        .form-group {\n            display: flex;\n            flex-direction: column;\n            margin-bottom: 20px;\n            width: 100%;\n        }\n\n        .form-group label {\n            font-size: 16px;\n            font-weight: 600;\n            margin-bottom: 10px;\n        }\n\n        .form-group input {\n            border: none;\n            outline: none;\n            font-size: 16px;\n            padding: 10px;\n            border-radius: 20px;\n            background-color: #f7f7f7;\n        }\n\n        .modal-buttons {\n            display: flex;\n            justify-content: space-between;\n            width: 100%;\n        }\n\n        .modal-buttons button {\n            border: none;\n            outline: none;\n            font-size: 16px;\n            font-weight: 600;\n            padding: 10px 20px;\n            cursor: pointer;\n            border-radius: 20px;\n            transition: all 0.3s ease;\n            margin: auto;\n        }\n\n        .modal-buttons button[type=\"submit\"] {\n            background-color: #007bff;\n            color: #fff;\n        }\n\n        .modal-buttons button[type=\"submit\"]:hover {\n            background-color: #0069d9;\n        }\n\n        .modal-buttons button[type=\"button\"] {\n            background-color: #f7f7f7;\n            color: #333;\n        }\n\n        .modal-buttons button[type=\"button\"]:hover {\n            background-color: #e5e5e5;\n        }\n\n        .chat-container {\n            display: flex;\n            flex-direction: column;\n            height: 100vh;\n        }\n\n        .chat-header {\n            background-color: #fff;\n            border-bottom: 1px solid #e5e5e5;\n            padding: 10px 20px;\n            display: flex;\n            justify-content: space-between;\n            align-items: center;\n        }\n\n        .chat-header h2 {\n            font-size: 20px;\n            font-weight: 600;\n            margin: 0;\n        }\n\n        .chat-messages {\n            flex-grow: 1;\n            overflow-y: scroll;\n            padding: 20px;\n        }\n\n        .message {\n            border-radius: 10px;\n            padding: 10px;\n            margin-bottom: 10px;\n            max-width: 70%;\n        }\n\n        .message {\n            background-color: #007bfc;\n            color: #fff;\n            align-self: flex-end;\n        }\n\n        .message a {\n            color: #fff;\n            text-decoration: underline;\n        }\n\n        .message a:hover {\n            color: #fff;\n            text-decoration: none;\n        }\n\n        .message .message-info {\n            display: flex;\n            justify-content: space-between;\n            align-items: center;\n            margin-bottom: 5px;\n        }\n\n        .message .message-info span {\n            font-size: 14px;\n            color: #777;\n        }\n\n        .message .message-info .sender {\n            font-weight: 600;\n            margin-right: 5px;\n        }\n\n        .message .message-info {\n            font-size: 12px;\n        }\n\n        .chat-form {\n            background-color: #fff;\n            border-top: 1px solid #e5e5e5;\n            padding: 20px;\n            display: flex;\n            align-items: center;\n        }\n\n        .chat-form input[type=\"text\"] {\n            border: none;\n            outline: none;\n            font-size: 16px;\n            padding: 10px;\n            border-radius: 20px;\n            background-color: #f7f7f7;\n            flex-grow: 1;\n            margin-right: 10px;\n        }\n\n        .chat-form button {\n            border: none;\n            outline: none;\n            font-size: 16px;\n            font-weight: 600;\n            padding: 10px 20px;\n            cursor: pointer;\n            border-radius: 20px;\n            transition: all 0.3s ease;\n            background-color: #007bff;\n            color: #fff;\n        }\n\n        .chat-form button:hover {\n            background-color: #0069d9;\n        }\n    </style>\n</head>\n\n<body>\n\n    <div class=\"overlay\" id=\"login-overlay\">\n        <div class=\"modal\">\n            <h3>登录</h3>\n            <form id=\"login-form\">\n                <div class=\"form-group\">\n                    <label for=\"username\">用户名</label>\n                    <input type=\"text\" id=\"username\">\n                </div>\n                <div class=\"form-group\">\n                    <label for=\"password\">密码</label>\n                    <input type=\"password\" id=\"password\">\n                </div>\n                <div class=\"modal-buttons\">\n                    <button type=\"submit\">登录</button>\n                </div>\n            </form>\n        </div>\n    </div>\n\n    <div class=\"chat-container\">\n        <div class=\"chat-header\">\n            <h2>聊天室</h2>\n        </div>\n\n        <div class=\"chat-messages\" id=\"chat-messages\">\n        </div>\n\n        <form class=\"chat-form\" id=\"chat-form\">\n            <input type=\"text\" id=\"chat-input\" placeholder=\"请输入消息...\">\n            <button type=\"submit\">发送</button>\n        </form>\n    </div>\n\n    <script>\n        var socket = new WebSocket(\"ws://\" + window.location.hostname + \":8888/ws\");\n        var login_name;\n        var login_pwd;\n        var chatroom;\n        var sendbool = false;\n        var login_data;\n        var json;\n        var chat_send = { \"username\": null, \"message\": null, \"token\": null };\n\n        const historyHttp = new XMLHttpRequest();\n        const url = '/history';\n        historyHttp.open(\"GET\", url);\n        historyHttp.send();\n        var s = 0;\n        historyHttp.onreadystatechange = (e) => {\n            console.log(historyHttp.responseText);\n            if (s == 1) {\n                chatroom = document.getElementById(\"chat-messages\");\n                chatroom.innerHTML += historyHttp.responseText;\n            }\n            s++;\n        }\n\n        socket.onmessage = function (event) {\n            var message = event.data;\n            chatroom = document.getElementById(\"chat-messages\");\n            chatroom.innerHTML += message;\n        };\n        document.getElementById(\"login-form\").addEventListener(\"submit\", function (event) {\n            var i = 1;\n            event.preventDefault();\n            if (login_name !== null && sendbool) {\n                alert(\"你tm登录过了\");\n            }\n            else {\n                login_name = document.getElementById(\"username\").value;\n                chat_send.username = login_name;\n                login_pwd = document.getElementById(\"password\").value;\n                const Http = new XMLHttpRequest();\n                const url = '/login?name=' + login_name + \"&pwd=\" + login_pwd;\n                Http.open(\"GET\", url);\n                Http.send();\n\n                Http.onreadystatechange = (e) => {\n                    login_data = Http.responseText;\n                    console.log(i);\n                    if (login_data !== \"\" && i == 1) {\n                        i++;\n                        json = JSON.parse(decodeURIComponent(login_data));\n                        console.log(json);\n                        chat_send.token = json.uuid;\n                        if (json.err === null) {\n                            document.getElementById(\"login-overlay\").style.display = \"none\";\n                            sendbool = true;\n                        }\n                        else {\n                            alert(json.err);\n                        }\n                    }\n                }\n            }\n        })\n        document.getElementById(\"chat-form\").addEventListener(\"submit\", function (event) {\n            event.preventDefault();\n            var messageInput = document.getElementById(\"chat-input\");\n            var message = messageInput.value;\n            if (login_name !== null && sendbool && chat_send.token !== null) {\n                chat_send.message = message;\n                socket.send(JSON.stringify(chat_send));\n                chat_send.message = \"\";\n            }\n            else {\n                alert(\"你tm给我登录\");\n            }\n            messageInput.value = \"\";\n        });\n    </script>\n</body>\n\n</html>"

func LocalIPv4s() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}

	return ips, nil
}

// PathExists 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		// 创建文件夹
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Printf("mkdir failed![%v]\n", err)
		} else {
			return true, nil
		}
	}
	return false, err
}

func init() {
	_, err := os.Lstat("./users.json")
	_, errh := os.Lstat("./public/index.html")
	if err != nil || errh != nil {
		if errh != nil {
			PathExists("public")
			fh, err := os.Create("./public/index.html")
			if err != nil {
				log.Println(err)
			}
			_, err = fh.Write([]byte(h))
			if err != nil {
				log.Println(err)
			}
			return
		}
		defer os.Exit(100)
		f, err := os.Create("./users.json")
		if err != nil {
			log.Println("创建配置文件错误：", err)
		}
		_, err = f.Write([]byte(d))
		if err != nil {

			log.Println("写入默认配置错误：", err)
		}
		log.Println("请填写配置文件: " + "./users.json")
	}
	data, err := os.ReadFile("./users.json")
	if err != nil {
		log.Println("配置文件读取失败：", err)
		os.Exit(200)
	}
	json := gjson.Get(string(data), "users").Map()
	users = make([]*user, len(json)+1)
	var i = 1
	users[0] = &user{name: "yxlr", pwd: "1145141919810homo"}
	for v, s := range json {
		users[i] = &user{name: v, pwd: s.String()}
		i++
	}
}

func handleWebSocket(c *gin.Context) {
	// 升级HTTP连接为WebSocket协议
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// 将新连接添加到客户端列表中
	clients[conn] = true

	// 无限循环读取从客户端发来的消息
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(clients, conn)
			return
		}
		x, err := tokenBool(p)
		if err != nil {
			log.Println(err)
			continue
		}

		// 将消息广播给所有连接的客户端
		for c := range clients {
			err := c.WriteMessage(messageType, x)
			if err != nil {
				log.Println(err)
				delete(clients, c)
				return
			}
		}
	}
}

func serveHome(c *gin.Context) {
	// 渲染HTML模板
	c.HTML(200, "index.html", gin.H{
		"history": history,
	})
}

func login(c *gin.Context) {
	userName := c.Query("name")
	userPwd := c.Query("pwd")
	if userName == "" || userPwd == "" {
		c.JSON(200, gin.H{
			"err":  "密码或账号为空",
			"uuid": nil,
		})
		return
	}
	for i, v := range users {
		if userName == v.name && userPwd == v.pwd {
			users[i].token = uuid.New().String()

			c.JSON(200, gin.H{
				"err":  nil,
				"uuid": users[i].token,
			})
			return
		}
	}
	c.JSON(200, gin.H{
		"err":  "账号或密码错误",
		"uuid": nil,
	})
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// 挂载WebSocket处理函数
	router.GET("/ws", handleWebSocket)

	// 挂载HTML文件
	router.LoadHTMLGlob("public/*")
	router.GET("/", serveHome)

	router.GET("/login", login)

	router.GET("/history", func(c *gin.Context) {
		c.String(200, history)
	})

	router.GET("/clear", func(c *gin.Context) {
		history = ""
		c.String(200, "聊天记录已全部删除")
	})

	addrs , err := LocalIPv4s()
	if err != nil {
		log.Println("获取本机ip失败",err)
	}
	if len(addrs) == 0 {
		log.Println("没有找到本机ip")
	}
	log.Println("running in ",addrs,"……")
	err = router.Run(":8888")
	if err != nil {
		log.Fatal(err)
	}
}

func tokenBool(s []byte) ([]byte, error) {
	var data *userMessage
	var p string
	err := json.Unmarshal(s, &data)
	log.Println(data)
	if err != nil {
		goto erraaa
	}
	for _, b := range users {
		if b.token == data.Token {
			p = fmt.Sprintf("<div class=\"message\"><div class=\"message-info\"><span class=\"sender\">%v</span></div><p>%v</p></div>", data.Username, data.Message)
			history += p
			return []byte(p), nil
		}
	}
erraaa:
	p = fmt.Sprintf("<div class=\"message\"><div class=\"message-info\"><span class=\"sender\">%v</span></div><p>有人使用了错误的token或使用奇奇怪怪的方法发送消息，信息为：%v</p></div>", "系统", string(s))
	history += p
	return []byte(p), nil
}
