package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// NetAssistantApp Main
type NetAssistantApp struct {
	receCount int
	sendCount int

	chanClose chan bool
	chanData  chan string
	listener  net.Listener
	connList  []net.Conn

	appWindow             *gtk.ApplicationWindow // app 主窗口
	combProtoType         *gtk.ComboBoxText      // 服务类型下拉框
	entryIP               *gtk.Entry             // IP地址
	entryPort             *gtk.Entry             // 端口
	btnConnect            *gtk.Button            // 连接按钮
	btnClearRecvDisplay   *gtk.Button            // 清空接收区
	btnClearSendDisplay   *gtk.Button            // 清空发送区
	labelStatus           *gtk.Label             // 当前状态提示
	labelSendCount        *gtk.Label             // 发送计数
	labelReceveCount      *gtk.Label             // 接收计数
	btnCleanCount         *gtk.Button            // 复位计数按钮
	tvDataReceive         *gtk.TextView          // 数据接收区
	swDataRec             *gtk.ScrolledWindow    // 数据接收区的滚动窗口
	tvDataSend            *gtk.TextView          // 数据发送区
	btnSend               *gtk.Button            // 发送消息按钮
	entryCurAddr          *gtk.Entry             // 当前监听地址
	entryCurPort          *gtk.Entry             // 当前监听端口
	cbHexDisplay          *gtk.CheckButton       // 16进制显示
	cbPauseDisplay        *gtk.CheckButton       // 暂停显示
	cbDisplayDate         *gtk.CheckButton       // 接收显示日期且换行
	cbDataSourceCycleSend *gtk.CheckButton       // 数据循环发送
	cbSendByHex           *gtk.CheckButton       // 数据以16进制发送
	tbReceData            *gtk.TextBuffer        //接收区buffer
	tbSendData            *gtk.TextBuffer        // 发送区buffer
	entryCycleTime        *gtk.Entry             // 持续发送数据的间隔
	cbAutoCleanAfterSend  *gtk.CheckButton       // 发送后清空
}

// NetAssistantAppNew create new instance
func NetAssistantAppNew() *NetAssistantApp {
	obj := &NetAssistantApp{}
	obj.chanClose = make(chan bool)
	obj.chanData = make(chan string)
	return obj
}

func (app *NetAssistantApp) update(recvStr string) {
	list := []string{}
	if app.cbHexDisplay.GetActive() {
		for i := 0; i < len(recvStr); i++ {
			log.Println(i, recvStr[i])
			list = append(list, fmt.Sprintf("%X", recvStr[i]))
		}
		recvStr = strings.Join(list, " ")
	}

	if app.cbDisplayDate.GetActive() {
		recvStr = fmt.Sprintf("[%s]:%s\n", time.Now().Format("2006-01-02 15:04:05"), recvStr)
	}
	iter := app.tbReceData.GetEndIter()
	app.tbReceData.Insert(iter, recvStr)
	app.labelReceveCount.SetText("接收计数：" + strconv.Itoa(app.receCount))
	app.tbReceData.CreateMark("end", iter, false)
	mark := app.tbReceData.GetMark("end")

	app.tvDataReceive.ScrollMarkOnscreen(mark)
}

func (app *NetAssistantApp) updateSendCount(count int) {
	app.sendCount += count
	app.labelSendCount.SetText("发送计数：" + strconv.Itoa(app.sendCount))
}

func (app *NetAssistantApp) handler(conn net.Conn) {
	defer conn.Close() // 关闭连接

	for {
		reader := bufio.NewReader(conn)
		var buf [2048]byte
		n, err := reader.Read(buf[:]) // 读取数据
		if err != nil {
			fmt.Println("从客户端读取数据异常，关闭此连接:", err)
			ss := conn.RemoteAddr().String()
			tips := fmt.Sprintf(`<span foreground="pink">😄 connection close: %s </span>`, ss)
			app.labelStatus.SetMarkup(tips)
			for index, connItem := range app.connList {
				if conn.LocalAddr().String() == connItem.LocalAddr().String() {
					app.connList = append(app.connList[:index], app.connList[index+1:]...)
					log.Println("已经将无效的连接移除")
				}
			}
			return
		}
		app.receCount += n
		recvStr := string(buf[:n])
		if !app.cbPauseDisplay.GetActive() {
			glib.IdleAdd(app.update, recvStr) //Make sure is running on the gui thread.
		}

	}
}

func (app *NetAssistantApp) createTCPClient(address string) (net.Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	strAddr := conn.LocalAddr().String()
	arr := strings.Split(strAddr, ":")
	app.entryCurPort.SetText(arr[1])
	app.entryCurAddr.SetText(arr[0])

	go app.handler(conn)
	return conn, nil
}

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}

func (app *NetAssistantApp) createTCPServer(addr string) (net.Listener, error) {
	listen, err := net.Listen("tcp", addr)

	if err != nil {
		log.Println("listen failed, err:", err)
		return nil, err
	}

	go func() {
		for {
			conn, err := listen.Accept() // 等待客户端
			if err != nil {
				log.Println("accept 失败, err:", err, "退出监听")
				return
			}
			ss := conn.RemoteAddr().String()
			tips := fmt.Sprintf(`<span foreground="green">😄 New connection: %s </span>`, ss)
			app.labelStatus.SetMarkup(tips)
			app.connList = append(app.connList, conn)
			go app.handler(conn)
		}

	}()

	return listen, nil
}

func (app *NetAssistantApp) onBtnCleanCount() {
	app.receCount = 0
	app.sendCount = 0
	app.labelReceveCount.SetText("接收计数：0")
	app.labelSendCount.SetText("发送计数：0")
}

func (app *NetAssistantApp) connectOrDisconectServer(isDisconnect bool, host, port string) {
	if !isDisconnect {
		log.Println("开始监听")
		if app.listener != nil {
			app.listener.Close()
		}
		listener, err := app.createTCPServer(host + ":" + port)

		if err != nil {
			strTips := fmt.Sprintf(`<span foreground="red">😱 %s</span>`, err)
			app.labelStatus.SetMarkup(strTips)
		} else {
			app.listener = listener
			strTips := `<span size="x-large" foreground="green">😄</span>`
			app.labelStatus.SetMarkup(strTips)
			app.btnConnect.SetLabel("Disconnect")
			app.entryCurPort.SetText(port)
			app.entryCurAddr.SetText(host)
			app.combProtoType.SetSensitive(false)
		}
	} else {
		log.Println("断开监听")
		if app.listener != nil {
			app.listener.Close()
			app.listener = nil
			for _, conn := range app.connList {
				conn.Close()
			}
			app.connList = []net.Conn{}
			fmt.Println("清空连接")

		}
		strTips := `<span foreground="green" size="x-large" >😎</span>`
		app.labelStatus.SetMarkup(strTips)
		app.btnConnect.SetLabel("Connect")
		app.entryCurAddr.SetText("")
		app.entryCurPort.SetText("")
		app.combProtoType.SetSensitive(true)
	}

}

func (app *NetAssistantApp) connectOrDisconectClient(isDisconnect bool, host, port string) {
	if !isDisconnect {
		log.Println("连接客户端")
		if len(app.connList) != 0 {
			for _, conn := range app.connList {
				conn.Close()
			}
			app.connList = []net.Conn{}
		}
		conn, err := app.createTCPClient(host + ":" + port)

		if err != nil {
			strTips := fmt.Sprintf(`<span foreground="red">😱 %s</span>`, err)
			app.labelStatus.SetMarkup(strTips)
		} else {
			app.connList = append(app.connList, conn)
			strTips := `<span size="x-large" foreground="green">😄</span>`
			app.labelStatus.SetMarkup(strTips)
			app.btnConnect.SetLabel("Disconnect")
			ss := conn.LocalAddr().String()
			ssArr := strings.Split(ss, ":")
			app.entryCurAddr.SetText(ssArr[0])
			app.entryCurPort.SetText(ssArr[1])
			app.combProtoType.SetSensitive(false)
		}
	} else {
		log.Println("断开客户端")

		for _, conn := range app.connList {
			conn.Close()
		}
		app.connList = []net.Conn{}
		log.Println("清空连接")

		strTips := `<span foreground="green" size="x-large" >😎</span>`
		app.labelStatus.SetMarkup(strTips)
		app.btnConnect.SetLabel("Connect")
		app.entryCurAddr.SetText("")
		app.entryCurPort.SetText("")
		app.combProtoType.SetSensitive(true)
	}
}

func (app *NetAssistantApp) onBtnConnect(button *gtk.Button) {

	strIP, _ := app.entryIP.GetText()
	strPort, _ := app.entryPort.GetText()
	serverType := app.combProtoType.GetActive()

	label, _ := app.btnConnect.GetLabel()
	isDisconnect := label == "Disconnect"
	if serverType == 1 {
		app.connectOrDisconectServer(isDisconnect, strIP, strPort)

	} else if serverType == 0 {
		app.connectOrDisconectClient(isDisconnect, strIP, strPort)
	}

}

func (app *NetAssistantApp) onBtnSend() {

	buff, err := app.tvDataSend.GetBuffer()
	if err != nil {
		fmt.Println(err)
	}

	start, end := buff.GetBounds()
	data, _ := buff.GetText(start, end, true)

	sendData := []byte(data)

	if app.cbSendByHex.GetActive() {
		data = strings.Replace(data, " ", "", -1)
		data = strings.Replace(data, "\n", "", -1)
		hexData, err := hex.DecodeString(data)
		if err != nil {
			log.Println(err)
			strTips := fmt.Sprintf(`<span foreground="red">😱 %s</span>`, err)
			app.labelStatus.SetMarkup(strTips)
		} else {
			sendData = hexData
		}
		log.Println(hexData)
	}

	label, err := app.btnSend.GetLabel()
	if label != "Send" {
		app.chanClose <- true
		app.btnSend.SetLabel("Send")
		return
	}

	if app.cbDataSourceCycleSend.GetActive() {
		// 数据是否循环发送
		app.btnSend.SetLabel("Stop")
		strCycleTime, err := app.entryCycleTime.GetText()
		if err != nil {
			strCycleTime = "1000"
		}
		cycle, err := strconv.Atoi(strCycleTime)
		if err != nil {
			cycle = 1000
		}
		go func(cycleTime int) {
		END:
			for {
				select {
				case <-app.chanClose:
					break END
				default:
					for _, conn := range app.connList {
						conn.Write(sendData)
						fmt.Println("Write data", data)
						app.updateSendCount(len(sendData))
					}
					if len(app.connList) == 0 {

						glib.IdleAdd(func() {
							app.labelStatus.SetText("没有TCP连接了")
							app.btnSend.SetLabel("Send")
						})
						break END

					}
				}
				time.Sleep(time.Duration(cycleTime) * time.Millisecond)
			}

		}(cycle)
	} else {

		for _, conn := range app.connList {
			conn.Write(sendData)
			fmt.Println("Write data", data)
			app.updateSendCount(len(sendData))
		}

	}

	if app.cbAutoCleanAfterSend.GetActive() {
		buff.SetText("")
	}

}

func (app *NetAssistantApp) onBtnClearRecvDisplay() {
	app.tbReceData.SetText("")

}

func (app *NetAssistantApp) doActivate(application *gtk.Application) {
	app.appWindow, _ = gtk.ApplicationWindowNew(application)
	app.appWindow.SetPosition(gtk.WIN_POS_CENTER)
	app.appWindow.SetResizable(false)
	app.appWindow.SetIconFromFile("./icon.png")

	app.appWindow.SetBorderWidth(10)
	app.appWindow.SetTitle("NetAssistant")

	// 总体容器
	windowContainer, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	windowContainerMiddle, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	windowContainerLeft, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	windowContainerRight, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	windowContainerBottom, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)

	// 左边的布局
	frame, _ := gtk.FrameNew("网络设置")
	frame.SetLabelAlign(0.1, 0.5)
	verticalBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	// 1 服务类型的组件
	labelProtType, _ := gtk.LabelNew("服务类型")
	labelProtType.SetXAlign(0)
	app.combProtoType, _ = gtk.ComboBoxTextNew()
	app.combProtoType.AppendText("TCP Client")
	app.combProtoType.AppendText("TCP Server")
	app.combProtoType.SetActive(0)
	// 添加到容器
	verticalBox.PackStart(labelProtType, false, false, 0)
	verticalBox.PackStart(app.combProtoType, false, false, 0)
	// 2 服务器IP设置
	labelIP, _ := gtk.LabelNew("IP设置")
	labelIP.SetXAlign(0)
	app.entryIP, _ = gtk.EntryNew()
	app.entryIP.SetText("127.0.0.1")
	verticalBox.PackStart(labelIP, false, false, 0)
	verticalBox.PackStart(app.entryIP, false, false, 0)
	// 3 服务器端口
	labelPort, _ := gtk.LabelNew("端口设置")
	labelPort.SetXAlign(0)
	app.entryPort, _ = gtk.EntryNew()
	app.entryPort.SetText("8003")
	verticalBox.PackStart(labelPort, false, false, 0)
	verticalBox.PackStart(app.entryPort, false, false, 0)
	// 4 连接按钮
	app.btnConnect, _ = gtk.ButtonNewWithLabel("Connect")
	app.btnConnect.Connect("clicked", app.onBtnConnect)
	verticalBox.PackStart(app.btnConnect, false, false, 0)

	// 5 两个切换按钮
	notebookTab, _ := gtk.NotebookNew()
	label1, _ := gtk.LabelNew("接收设置")
	label2, _ := gtk.LabelNew("发送设置")

	// 接收设置
	frame1ContentBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	cbReceive2File, _ := gtk.CheckButtonNewWithLabel("接收转向文件")
	app.cbDisplayDate, _ = gtk.CheckButtonNewWithLabel("显示时间且换行")
	app.cbHexDisplay, _ = gtk.CheckButtonNewWithLabel("十六进制显示")
	app.cbPauseDisplay, _ = gtk.CheckButtonNewWithLabel("暂停接收")
	btnHboxContainer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	saveDataCb, _ := gtk.ButtonNewWithLabel("保存数据")
	app.btnClearRecvDisplay, _ = gtk.ButtonNewWithLabel("清空显示")
	app.btnClearRecvDisplay.Connect("clicked", app.onBtnClearRecvDisplay)

	btnHboxContainer.PackStart(saveDataCb, true, false, 0)
	btnHboxContainer.PackStart(app.btnClearRecvDisplay, true, false, 0)
	frame1ContentBox.PackStart(cbReceive2File, false, false, 0)
	frame1ContentBox.PackStart(app.cbDisplayDate, false, false, 0)
	frame1ContentBox.PackStart(app.cbHexDisplay, false, false, 0)
	frame1ContentBox.PackStart(app.cbPauseDisplay, false, false, 0)
	frame1ContentBox.PackStart(btnHboxContainer, false, false, 0)
	frame1ContentBox.SetBorderWidth(10)

	// 发送设置
	frame2ContentBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	enabelFileSourceCb, _ := gtk.CheckButtonNewWithLabel("启用文件数据源发送")
	app.cbAutoCleanAfterSend, _ = gtk.CheckButtonNewWithLabel("发送完自动清空")
	app.cbSendByHex, _ = gtk.CheckButtonNewWithLabel("按十六进制发送")
	app.cbDataSourceCycleSend, _ = gtk.CheckButtonNewWithLabel("数据源循环发送")
	app.entryCycleTime, _ = gtk.EntryNew()
	app.entryCycleTime.SetPlaceholderText("间隔毫秒，默认1000")
	btnHboxContainer2, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	loadDataBtn, _ := gtk.ButtonNewWithLabel("加载数据")
	app.btnClearSendDisplay, _ = gtk.ButtonNewWithLabel("清空显示")

	frame2ContentBox.PackStart(enabelFileSourceCb, false, false, 0)
	frame2ContentBox.PackStart(app.cbAutoCleanAfterSend, false, false, 0)
	frame2ContentBox.PackStart(app.cbSendByHex, false, false, 0)
	frame2ContentBox.PackStart(app.cbDataSourceCycleSend, false, false, 0)
	frame2ContentBox.PackStart(app.entryCycleTime, false, false, 0)
	btnHboxContainer2.PackStart(loadDataBtn, true, false, 0)
	btnHboxContainer2.PackStart(app.btnClearSendDisplay, true, false, 0)
	frame2ContentBox.PackStart(btnHboxContainer2, false, false, 0)
	frame2ContentBox.SetBorderWidth(10)

	frame1, _ := gtk.FrameNew("") // 接收设置的frame
	frame1.Add(frame1ContentBox)
	frame2, _ := gtk.FrameNew("") // 发送设置的frame
	frame2.Add(frame2ContentBox)
	// notebookTab.Add(label1)
	notebookTab.AppendPage(frame1, label1)
	notebookTab.AppendPage(frame2, label2)

	// 右边的布局
	titleDataReceiveArea, _ := gtk.LabelNew("数据接收区")
	titleDataReceiveArea.SetXAlign(0)
	windowContainerRight.PackStart(titleDataReceiveArea, false, false, 0)
	app.swDataRec, _ = gtk.ScrolledWindowNew(nil, nil)
	app.tvDataReceive, _ = gtk.TextViewNew()
	app.tvDataReceive.SetEditable(false)
	app.tvDataReceive.SetWrapMode(gtk.WRAP_CHAR)
	app.swDataRec.Add(app.tvDataReceive)
	windowContainerRight.PackStart(app.swDataRec, true, true, 0)
	middleContainer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	labelLocalAddr, _ := gtk.LabelNew("当前地址")
	app.entryCurAddr, _ = gtk.EntryNew()
	app.entryCurAddr.SetEditable(false)
	labelLocalPort, _ := gtk.LabelNew("当前端口")
	app.entryCurPort, _ = gtk.EntryNew()
	app.entryCurPort.SetEditable(false)
	middleContainer.PackStart(labelLocalAddr, false, false, 0)
	middleContainer.PackStart(app.entryCurAddr, false, false, 0)
	middleContainer.PackStart(labelLocalPort, false, false, 0)
	middleContainer.PackStart(app.entryCurPort, false, false, 0)
	windowContainerRight.PackStart(middleContainer, false, false, 0)
	bottomContainer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	scrollerDataSend, _ := gtk.ScrolledWindowNew(nil, nil)
	app.tvDataSend, _ = gtk.TextViewNew()
	app.tvDataSend.SetWrapMode(gtk.WRAP_CHAR)

	scrollerDataSend.Add(app.tvDataSend)
	scrollerDataSend.SetSizeRequest(-1, 180)
	boxSendBtn, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	app.btnSend, _ = gtk.ButtonNewWithLabel("Send")
	app.btnSend.Connect("clicked", app.onBtnSend)
	boxSendBtn.PackEnd(app.btnSend, false, false, 0)
	app.btnSend.SetSizeRequest(80, -1)
	bottomContainer.PackStart(scrollerDataSend, true, true, 0)
	bottomContainer.PackEnd(boxSendBtn, false, false, 0)
	windowContainerRight.PackStart(bottomContainer, false, false, 0)

	// 最底下的布局
	app.labelStatus, _ = gtk.LabelNew("")
	app.labelStatus.SetMarkup(`<span foreground="green" size="x-large" >😎</span>`)
	windowContainerBottom.PackStart(app.labelStatus, true, false, 0)
	app.labelSendCount, _ = gtk.LabelNew("发送计数 0")
	windowContainerBottom.PackStart(app.labelSendCount, true, false, 0)
	app.labelReceveCount, _ = gtk.LabelNew("接收计数 0")
	windowContainerBottom.PackStart(app.labelReceveCount, true, false, 0)
	app.btnCleanCount, _ = gtk.ButtonNewWithLabel("复位计数")
	app.btnCleanCount.Connect("clicked", app.onBtnCleanCount)

	windowContainerBottom.PackEnd(app.btnCleanCount, false, false, 0)

	frame.Add(verticalBox)
	app.appWindow.Add(windowContainer) // 垂直布局

	windowContainerLeft.PackStart(frame, false, false, 0)
	windowContainerLeft.PackStart(notebookTab, false, false, 0)
	windowContainerMiddle.PackStart(windowContainerLeft, false, false, 0)
	windowContainerMiddle.PackStart(windowContainerRight, false, false, 0)

	windowContainer.PackStart(windowContainerMiddle, false, false, 0)
	windowContainer.PackStart(windowContainerBottom, false, false, 0)

	app.appWindow.SetDefaultSize(400, 400)
	app.appWindow.ShowAll()

	if app.tbReceData == nil {
		app.tbReceData, _ = gtk.TextBufferNew(nil)
		app.tvDataReceive.SetBuffer(app.tbReceData)
	}
}

func main() {

	const appID = "org.gtk.example"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_NON_UNIQUE)

	if err != nil {
		log.Fatal("Could not create application.", err)
	}
	app := NetAssistantAppNew()
	application.Connect("activate", app.doActivate)

	application.Run(os.Args)
}
