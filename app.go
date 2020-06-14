package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// NetAssistantApp Main
type NetAssistantApp struct {
	receCount int
	sendCoutn int

	chanClose chan bool
	chanData  chan string

	appWindow           *gtk.ApplicationWindow // app 主窗口
	comb                *gtk.ComboBoxText      // 服务类型下拉框
	entryIP             *gtk.Entry             // IP地址
	entryPort           *gtk.Entry             // 端口
	buttonConnect       *gtk.Button            // 连接按钮
	clearReceDisplayCb  *gtk.Button            // 清空接收区
	clearSendDisplayBtn *gtk.Button            // 清空发送区
	labelStatus         *gtk.Label             // 当前状态提示
	labelSendCount      *gtk.Label             // 发送计数
	labelReceveCount    *gtk.Label             // 接收计数
	btnCleanCount       *gtk.Button            // 复位计数按钮
	textViewDataReceive *gtk.TextView          // 数据接收区
	scrollerDataRec     *gtk.ScrolledWindow
	textViewDataSend    *gtk.TextView // 数据发送区
	sendBtn             *gtk.Button   // 发送消息按钮
	entryLocalAddr      *gtk.Entry    // 当前监听地址
	entryLocalPort      *gtk.Entry    // 当前监听端口

	bufferRecevData *gtk.TextBuffer
	bufferSendData  *gtk.TextBuffer
}

// NetAssistantAppNew create new instance
func NetAssistantAppNew() *NetAssistantApp {
	obj := &NetAssistantApp{}
	obj.chanClose = make(chan bool)
	obj.chanData = make(chan string)
	return obj
}

func (app *NetAssistantApp) update(recvStr string) {
	iter := app.bufferRecevData.GetEndIter()
	app.bufferRecevData.Insert(iter, recvStr)
	app.labelReceveCount.SetText("接收计数：" + strconv.Itoa(app.receCount))
	app.bufferRecevData.CreateMark("end", iter, false)
	mark := app.bufferRecevData.GetMark("end")
	app.textViewDataReceive.ScrollMarkOnscreen(mark)
}

func (app *NetAssistantApp) process(conn net.Conn) {
	defer conn.Close() // 关闭连接
	go func() {
		for {
			select {
			case value, _ := <-app.chanData:
				conn.Write([]byte(value))
				fmt.Println("发送了", value)
			}
		}
	}()
	for {
		reader := bufio.NewReader(conn)
		var buf [2048]byte
		n, err := reader.Read(buf[:]) // 读取数据
		if err != nil {
			fmt.Println("read from client failed, err:", err)
			break
		}
		app.receCount += n
		recvStr := string(buf[:n])
		fmt.Println("收到client端发来的数据：", recvStr)
		fmt.Println("解析数据1")
		if app.bufferRecevData == nil {
			app.bufferRecevData, _ = gtk.TextBufferNew(nil)
			fmt.Println("解析数据2")
			app.textViewDataReceive.SetBuffer(app.bufferRecevData)
			fmt.Println("解析数据3")
		}

		glib.IdleAdd(app.update, recvStr) //Make sure is running on the gui thread.
	}
}

func (app *NetAssistantApp) createTCPClient(address string) error {
	conn, err := net.Dial("tcp", address)
	go func() {
		for {
			select {
			case value := <-app.chanClose:
				fmt.Println("app.chanClose", value)
				conn.Close()

				fmt.Println("关闭主连接")
				return
			}
		}
	}()
	if err != nil {
		return err
	}
	go app.process(conn)
	return nil
}

func (app *NetAssistantApp) createTCPServer(host string, port int) error {
	listen, err := net.Listen("tcp", host+":"+strconv.Itoa(port))

	if err != nil {
		fmt.Println("listen failed, err:", err)
		return err
	}

	go func() {
		for {
			select {
			case value := <-app.chanClose:
				fmt.Println("app.chanClose", value)
				listen.Close()

				fmt.Println("关闭主连接")
				return
			}
		}
	}()
	go func() {
		for {

			conn, err := listen.Accept() // 建立连接

			if err != nil {
				fmt.Println("accept failed, err:", err)
				return
			}

			ss := conn.RemoteAddr().String()
			tips := fmt.Sprintf(`<span foreground="green">😄 New Connect: %s </span>`, ss)
			app.labelStatus.SetMarkup(tips)

			go app.process(conn) // 启动一个goroutine处理连接
		}

	}()

	return nil
}

func (app *NetAssistantApp) onCleanCountClicked() {
	app.receCount = 0
	app.sendCoutn = 0
	app.labelReceveCount.SetText("接收计数：0")
	app.labelSendCount.SetText("发送计数：0")
}

func (app *NetAssistantApp) onConnectBtnClicked(button *gtk.Button) {
	strIP, _ := app.entryIP.GetText()
	strPort, _ := app.entryPort.GetText()
	serverType := app.comb.GetActive()
	port, err := strconv.Atoi(strPort)

	if err != nil {
		app.labelStatus.SetMarkup(`<span foreground="red">😰 Invalid Port</span>`)
		return
	}

	if serverType == 1 {
		label, _ := app.buttonConnect.GetLabel()
		if label == "Connect" {
			err = app.createTCPServer(strIP, port)
			if err != nil {
				strTips := fmt.Sprintf(`<span foreground="red">😱 %s</span>`, err)
				app.labelStatus.SetMarkup(strTips)
			} else {
				strTips := `<span size="x-large" foreground="green">😄</span>`
				app.labelStatus.SetMarkup(strTips)
				app.buttonConnect.SetLabel("Disconnect")
				app.entryLocalPort.SetText(strPort)
				app.entryLocalAddr.SetText(strIP)
				app.comb.SetSensitive(false)
			}
		} else {
			fmt.Println("断开连接")
			app.chanClose <- true
			strTips := `<span foreground="green" size="x-large" >😎</span>`
			app.labelStatus.SetMarkup(strTips)
			app.buttonConnect.SetLabel("Connect")
			app.entryLocalPort.SetText("")
			app.entryLocalAddr.SetText("")
			app.comb.SetSensitive(true)
		}
	} else if serverType == 0 {
		fmt.Println("创建客户端")
		label, _ := app.buttonConnect.GetLabel()
		if label == "Connect" {
			err := app.createTCPClient(strIP + ":" + strPort)
			if err != nil {
				strTips := fmt.Sprintf(`<span foreground="red">😱 %s</span>`, err)
				app.labelStatus.SetMarkup(strTips)
			} else {
				strTips := `<span size="x-large" foreground="green">😄</span>`
				app.labelStatus.SetMarkup(strTips)
				app.buttonConnect.SetLabel("Disconnect")
				app.entryLocalPort.SetText(strPort)
				app.entryLocalAddr.SetText(strIP)
				app.comb.SetSensitive(false)
			}
		} else {
			fmt.Println("断开连接Client")
			app.chanClose <- true
			strTips := `<span foreground="green" size="x-large" >😎</span>`
			app.labelStatus.SetMarkup(strTips)
			app.buttonConnect.SetLabel("Connect")
			app.entryLocalPort.SetText("")
			app.entryLocalAddr.SetText("")
			app.comb.SetSensitive(true)
		}
	}

}

func (app *NetAssistantApp) onSendMessageClicked() {
	buff, err := app.textViewDataSend.GetBuffer()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(buff)
	}

	start, end := buff.GetBounds()
	data, _ := buff.GetText(start, end, false)
	fmt.Println("data", data)
	go func() {
		app.chanData <- data
		fmt.Println("往app.sendData写入数据成功！")
	}()

}

func (app *NetAssistantApp) onClearRecvDisplay() {
	app.bufferRecevData.SetText("")
}

func (app *NetAssistantApp) doActivate(application *gtk.Application) {
	app.appWindow, _ = gtk.ApplicationWindowNew(application)
	app.appWindow.SetPosition(gtk.WIN_POS_CENTER)
	app.appWindow.SetResizable(false)
	app.appWindow.SetIconFromFile("./icon.png")

	app.appWindow.SetBorderWidth(10)
	app.appWindow.SetTitle("网络调试助手")

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
	app.comb, _ = gtk.ComboBoxTextNew()
	app.comb.AppendText("TCP Client")
	app.comb.AppendText("TCP Server")
	app.comb.SetActive(0)
	// 添加到容器
	verticalBox.PackStart(labelProtType, false, false, 0)
	verticalBox.PackStart(app.comb, false, false, 0)
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
	app.buttonConnect, _ = gtk.ButtonNewWithLabel("Connect")
	app.buttonConnect.Connect("clicked", app.onConnectBtnClicked)

	verticalBox.PackStart(app.buttonConnect, false, false, 0)

	// 5 两个切换按钮
	notebookTab, _ := gtk.NotebookNew()
	label1, _ := gtk.LabelNew("接收设置")
	label2, _ := gtk.LabelNew("发送设置")

	// 接收设置
	frame1ContentBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	receive2FileCb, _ := gtk.CheckButtonNewWithLabel("接收转向文件")
	displayDateCb, _ := gtk.CheckButtonNewWithLabel("显示接收日期")
	hexDisplayCb, _ := gtk.CheckButtonNewWithLabel("十六进制显示")
	pauseDisplayCb, _ := gtk.CheckButtonNewWithLabel("暂停接收显示")
	btnHboxContainer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	saveDataCb, _ := gtk.ButtonNewWithLabel("保存数据")
	app.clearReceDisplayCb, _ = gtk.ButtonNewWithLabel("清空显示")
	app.clearReceDisplayCb.Connect("clicked", app.onClearRecvDisplay)

	btnHboxContainer.PackStart(saveDataCb, true, false, 0)
	btnHboxContainer.PackStart(app.clearReceDisplayCb, true, false, 0)
	frame1ContentBox.PackStart(receive2FileCb, false, false, 0)
	frame1ContentBox.PackStart(displayDateCb, false, false, 0)
	frame1ContentBox.PackStart(hexDisplayCb, false, false, 0)
	frame1ContentBox.PackStart(pauseDisplayCb, false, false, 0)
	frame1ContentBox.PackStart(btnHboxContainer, false, false, 0)
	frame1ContentBox.SetBorderWidth(10)

	// 发送设置
	frame2ContentBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	enabelFileSourceCb, _ := gtk.CheckButtonNewWithLabel("启用文件数据源发送")
	autoCleanAfterSendCb, _ := gtk.CheckButtonNewWithLabel("发送完自动清空")
	sendByHexCb, _ := gtk.CheckButtonNewWithLabel("按十六进制发送")
	dataSourceCycleSendCb, _ := gtk.CheckButtonNewWithLabel("数据源循环发送")
	btnHboxContainer2, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	loadDataBtn, _ := gtk.ButtonNewWithLabel("加载数据")
	app.clearSendDisplayBtn, _ = gtk.ButtonNewWithLabel("清空显示")

	frame2ContentBox.PackStart(enabelFileSourceCb, false, false, 0)
	frame2ContentBox.PackStart(autoCleanAfterSendCb, false, false, 0)
	frame2ContentBox.PackStart(sendByHexCb, false, false, 0)
	frame2ContentBox.PackStart(dataSourceCycleSendCb, false, false, 0)
	btnHboxContainer2.PackStart(loadDataBtn, true, false, 0)
	btnHboxContainer2.PackStart(app.clearSendDisplayBtn, true, false, 0)
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
	app.scrollerDataRec, _ = gtk.ScrolledWindowNew(nil, nil)
	app.textViewDataReceive, _ = gtk.TextViewNew()
	app.textViewDataReceive.SetEditable(false)
	app.textViewDataReceive.SetWrapMode(gtk.WRAP_CHAR)
	app.scrollerDataRec.Add(app.textViewDataReceive)
	windowContainerRight.PackStart(app.scrollerDataRec, true, true, 0)
	middleContainer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	labelLocalAddr, _ := gtk.LabelNew("当前地址")
	app.entryLocalAddr, _ = gtk.EntryNew()
	app.entryLocalAddr.SetEditable(false)
	labelLocalPort, _ := gtk.LabelNew("当前端口")
	app.entryLocalPort, _ = gtk.EntryNew()
	app.entryLocalPort.SetEditable(false)
	middleContainer.PackStart(labelLocalAddr, false, false, 0)
	middleContainer.PackStart(app.entryLocalAddr, false, false, 0)
	middleContainer.PackStart(labelLocalPort, false, false, 0)
	middleContainer.PackStart(app.entryLocalPort, false, false, 0)
	windowContainerRight.PackStart(middleContainer, false, false, 0)
	bottomContainer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	scrollerDataSend, _ := gtk.ScrolledWindowNew(nil, nil)
	app.textViewDataSend, _ = gtk.TextViewNew()

	scrollerDataSend.Add(app.textViewDataSend)
	scrollerDataSend.SetSizeRequest(-1, 180)
	boxSendBtn, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	app.sendBtn, _ = gtk.ButtonNewWithLabel("发送")
	app.sendBtn.Connect("clicked", app.onSendMessageClicked)
	boxSendBtn.PackEnd(app.sendBtn, false, false, 0)
	app.sendBtn.SetSizeRequest(80, -1)
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
	app.btnCleanCount.Connect("clicked", app.onCleanCountClicked)

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
