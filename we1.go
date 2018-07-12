package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/axgle/mahonia"
	"golang.org/x/net/websocket"
)

var cmd string
var params []string

func sayHelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println(r.Form)
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])

	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello astaxie!")
}

func Login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.html")
		t.Execute(w, nil)
	} else {
		fmt.Println("cmd :", r.FormValue("cmd"))
		s := strings.Split(r.FormValue("cmd"), " ")
		cmd = s[0]
		fmt.Println("cmd :", s[0])
		params = s[1:]
		fmt.Println("Paraments :", params)
	}
}

/////////////////创建websocket
func Echo(ws *websocket.Conn) {
	var err error

	var replay string

	for {
		if err = websocket.Message.Receive(ws, &replay); err != nil {
			fmt.Println("Can't receive!")
			break
		}
		fmt.Println("Received back from client: " + replay)

		cmdStr := strings.Split(replay, " ")

		mycmd := exec.Command(cmdStr[0], cmdStr[1:]...)

		fmt.Println(mycmd.Args)

		stdout, err := mycmd.StdoutPipe()
		if err != nil {
			fmt.Println(err)
		}

		mycmd.Start()
		reader := bufio.NewReader(stdout)
		for {
			line, err2 := reader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}
			msg_str := convertToString(line, "gbk", "utf-8")
			fmt.Println("Sending to client: " + msg_str)
			if err = websocket.Message.Send(ws, msg_str); err != nil {
				fmt.Println("Can't send")
				break
			}
		}
		mycmd.Wait()
	}

}

////编码转换函数
func convertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

func main() {

	http.HandleFunc("/", sayHelloName)
	http.HandleFunc("/login", Login)
	http.Handle("/ws", websocket.Handler(Echo))

	err := http.ListenAndServe(":8088", nil)

	if err != nil {
		log.Fatal("ListenAndServer: ", err)
	}
}
