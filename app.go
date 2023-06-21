package main

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/op/go-logging"
)

const (
	icon                     = `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`
	IT_END            string = "end"
	IT_APP_NAME       string = "Network Assistant"
	IT_SETTING        string = "Settings"
	IT_TYPE           string = "Type"
	IT_TCP_CLIENT     string = "TCP Client"
	IT_TCP_SERVER     string = "TCP Server"
	IT_UDP_CLIENT     string = "UDP Client"
	IT_UDP_SERVER     string = "UDP Server"
	IT_PORT           string = "Port"
	IT_CONNECT        string = "Connect"
	IT_RECV_SETTINGS  string = "Recv Settings"
	IT_SEND_SETTINGS  string = "Send Settings"
	IT_SAVE_TO_FILE   string = "Save to file"
	IT_SHOW_RECV_TIME string = "Show recv time"
	IT_SHOW_HEX       string = "Show hex"
	IT_PAUSE          string = "Pause"
	IT_SAVE           string = "Save"
	IT_CLEAR          string = "Clear"
	IT_SEND           string = "Send"
	IT_APPEND_RN      string = "Append \\r\\n"
	IT_AUTO_CLEAR     string = "Auto clear"
	IT_SEND_HEX       string = "Send hex"
	IT_SEND_CIRC      string = "Send circularly"
	IT_LOAD_DATA      string = "Load data"
	IT_DATA_RECVED    string = "Data received"
	IT_WAIT_CONN      string = "Waiting connection"
	IT_SEND_COUNT     string = "Send count:"
	IT_RECEVER_COUNT  string = "Recv count:"
	IT_RESET          string = "Reset"
	IT_NO_CONN        string = "there's no connection"
	IT_DISCONNECT     string = "Disconnect"
	IT_LOCAL_IP       string = "Local ip"
	IT_LOCAL_PORT     string = "Local port"
	IT_STOP           string = "Stop"
)

var (
	log         = logging.MustGetLogger("APP")
	i18nTextMap = map[string]string{
		IT_END:            "结束",
		IT_APP_NAME:       "网络调试助手",
		IT_SETTING:        "设置",
		IT_TYPE:           "类型",
		IT_TCP_CLIENT:     "TCP客户端",
		IT_TCP_SERVER:     "TCP服务端",
		IT_UDP_CLIENT:     "UDP客户端",
		IT_UDP_SERVER:     "UDP服务端",
		IT_PORT:           "端口",
		IT_CONNECT:        "连接",
		IT_RECV_SETTINGS:  "接收设置",
		IT_SEND_SETTINGS:  "发送设置",
		IT_SAVE_TO_FILE:   "保存到文件",
		IT_SHOW_RECV_TIME: "显示接收时间",
		IT_SHOW_HEX:       "16进制显示",
		IT_PAUSE:          "暂停",
		IT_SAVE:           "保存",
		IT_CLEAR:          "清除",
		IT_SEND:           "发送",
		IT_APPEND_RN:      "追加\\r\\n",
		IT_AUTO_CLEAR:     "自动清除",
		IT_SEND_HEX:       "发送16进制",
		IT_SEND_CIRC:      "循环发送",
		IT_LOAD_DATA:      "加载数据",
		IT_DATA_RECVED:    "已接收数据",
		IT_WAIT_CONN:      "等待连接",
		IT_SEND_COUNT:     "发送数:",
		IT_RECEVER_COUNT:  "接收数:",
		IT_RESET:          "重置",
		IT_NO_CONN:        "没有连接",
		IT_DISCONNECT:     "断开连接",
		IT_LOCAL_IP:       "本地IP",
		IT_LOCAL_PORT:     "本地端口",
		IT_STOP:           "停止",
	}
	systemLangIsZh = strings.HasPrefix(os.Getenv("LANG"), "zh_")
)

func getI18nText(key string) string {
	if systemLangIsZh {
		if v, exist := i18nTextMap[key]; exist {
			return v
		}
	}
	return key

}

// NetAssistantApp Main
type NetAssistantApp struct {
	receCount int
	sendCount int

	chanClose chan bool
	listener  net.Listener
	connList  []net.Conn
	fileName  string

	appWindow             *gtk.ApplicationWindow
	combProtoType         *gtk.ComboBoxText
	entryIP               *gtk.Entry
	entryPort             *gtk.Entry
	btnConnect            *gtk.Button
	btnClearRecvDisplay   *gtk.Button
	btnClearSendDisplay   *gtk.Button
	labelStatus           *gtk.Label
	labelSendCount        *gtk.Label
	labelReceveCount      *gtk.Label
	btnCleanCount         *gtk.Button
	tvDataReceive         *gtk.TextView
	swDataRec             *gtk.ScrolledWindow
	tvDataSend            *gtk.TextView
	btnSend               *gtk.Button
	entryCurAddr          *gtk.Entry
	entryCurPort          *gtk.Entry
	cbHexDisplay          *gtk.CheckButton
	cbPauseDisplay        *gtk.CheckButton
	cbDisplayDate         *gtk.CheckButton
	cbDataSourceCycleSend *gtk.CheckButton
	cbSendByHex           *gtk.CheckButton
	tbReceData            *gtk.TextBuffer
	tbSendData            *gtk.TextBuffer
	entryCycleTime        *gtk.Entry
	cbAutoCleanAfterSend  *gtk.CheckButton
	cbReceive2File        *gtk.CheckButton
	btnSaveData           *gtk.Button
	btnLoadData           *gtk.Button
	labelLocalAddr        *gtk.Label
	labelLocalPort        *gtk.Label
	cbAppendNewLine       *gtk.CheckButton
}

// NetAssistantAppNew create new instance
func NetAssistantAppNew() *NetAssistantApp {
	obj := &NetAssistantApp{}
	obj.chanClose = make(chan bool)
	return obj
}

func appendConntent2File(filename string, content []byte) {
	fd, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	defer fd.Close()
	fd.Write(content)
}

func (app *NetAssistantApp) getRecvData() string {
	buff, err := app.tvDataReceive.GetBuffer()
	if err != nil {
		log.Error(err)
		return ""
	}

	start, end := buff.GetBounds()
	data, err := buff.GetText(start, end, true)
	if err != nil {
		log.Error(err)
		return ""
	}
	return data
}

func (app *NetAssistantApp) update(recvStr string) {
	list := []string{}
	if app.cbHexDisplay.GetActive() {
		for i := 0; i < len(recvStr); i++ {
			log.Debug(i, recvStr[i])
			list = append(list, fmt.Sprintf("%X", recvStr[i]))
		}
		recvStr = strings.Join(list, " ")
	}

	if app.cbDisplayDate.GetActive() {
		recvStr = fmt.Sprintf("[%s]:%s\n", time.Now().Format("2006-01-02 15:04:05"), recvStr)
	}

	if app.cbReceive2File.GetActive() {
		appendConntent2File(app.fileName, []byte(recvStr))
	}

	iter := app.tbReceData.GetEndIter()
	app.tbReceData.Insert(iter, recvStr)
	app.labelReceveCount.SetText(getI18nText(IT_RECEVER_COUNT) + strconv.Itoa(app.receCount))
	app.tbReceData.CreateMark(getI18nText(IT_END), iter, false)
	mark := app.tbReceData.GetMark(getI18nText(IT_END))
	app.tvDataReceive.ScrollMarkOnscreen(mark)
}

func (app *NetAssistantApp) updateSendCount(count int) {
	app.sendCount += count
	app.labelSendCount.SetText(getI18nText(IT_SEND_COUNT) + strconv.Itoa(app.sendCount))
}

func (app *NetAssistantApp) handler(conn net.Conn) {
	defer conn.Close() // close connection
	for {
		reader := bufio.NewReader(conn)
		var buf [2048]byte
		n, err := reader.Read(buf[:])
		if err != nil {
			log.Info("close connection:", err)
			_, ok := conn.(*net.UDPConn)
			if !ok {
				ss := conn.RemoteAddr().String()
				tips := fmt.Sprintf(`<span foreground="pink">connection closed: %s </span>`, ss)
				glib.IdleAdd(func() {
					app.labelStatus.SetMarkup(tips)
				})
			}
			for index, connItem := range app.connList {
				if conn.LocalAddr().String() == connItem.LocalAddr().String() {
					app.connList = append(app.connList[:index], app.connList[index+1:]...)
				}
			}
			return
		}
		app.receCount += n
		recvStr := string(buf[:n])
		if !app.cbPauseDisplay.GetActive() {
			glib.IdleAdd(func() {
				list := []string{}
				if app.cbHexDisplay.GetActive() {
					for i := 0; i < len(recvStr); i++ {
						log.Debug(i, recvStr[i])
						list = append(list, fmt.Sprintf("%X", recvStr[i]))
					}
					recvStr = strings.Join(list, " ")
				}

				if app.cbDisplayDate.GetActive() {
					recvStr = fmt.Sprintf("[%s]:%s\n", time.Now().Format("2006-01-02 15:04:05"), recvStr)
				}

				if app.cbReceive2File.GetActive() {
					appendConntent2File(app.fileName, []byte(recvStr))
				}

				iter := app.tbReceData.GetEndIter()
				app.tbReceData.Insert(iter, recvStr)
				app.labelReceveCount.SetText(getI18nText(IT_RECEVER_COUNT) + strconv.Itoa(app.receCount))
				app.tbReceData.CreateMark(getI18nText(IT_END), iter, false)
				mark := app.tbReceData.GetMark(getI18nText(IT_END))
				app.tvDataReceive.ScrollMarkOnscreen(mark)
			}) //Make sure is running on the gui thread.
		}
	}
}

func (app *NetAssistantApp) onBtnCleanCount() {
	app.receCount = 0
	app.sendCount = 0
	app.labelReceveCount.SetText(getI18nText(IT_RECEVER_COUNT))
	app.labelSendCount.SetText(getI18nText(IT_SEND_COUNT))
}

func (app *NetAssistantApp) onCbReceive2File() {
	if app.cbReceive2File.GetActive() {
		dialog, _ := gtk.FileChooserNativeDialogNew("Select File", app.appWindow, gtk.FILE_CHOOSER_ACTION_OPEN, "Select", "Cancel")
		res := dialog.Run()
		if res == int(gtk.RESPONSE_ACCEPT) {
			fileName := dialog.FileChooser.GetFilename()
			app.fileName = fileName
		}
		dialog.Destroy()
	}
}

func (app *NetAssistantApp) onBtnLoadData() {
	dialog, _ := gtk.FileChooserNativeDialogNew("Select File", app.appWindow, gtk.FILE_CHOOSER_ACTION_OPEN, "Select", "Cancel")
	res := dialog.Run()
	if res == int(gtk.RESPONSE_ACCEPT) {
		fileName := dialog.FileChooser.GetFilename()
		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Error(err)
		} else {
			buf, _ := app.tvDataSend.GetBuffer()
			buf.SetText(string(data))
		}
	}
	dialog.Destroy()
}

func (app *NetAssistantApp) onBtnSaveData() {
	dialog, _ := gtk.FileChooserNativeDialogNew(getI18nText(IT_SAVE_TO_FILE), app.appWindow, gtk.FILE_CHOOSER_ACTION_SAVE, "Save", "Cancel")
	res := dialog.Run()
	if res == int(gtk.RESPONSE_ACCEPT) {
		fileName := dialog.FileChooser.GetFilename()
		appendConntent2File(fileName, []byte(app.getRecvData()))
	}
	dialog.Destroy()
}

func (app *NetAssistantApp) addConnection(conn net.Conn) {
	app.connList = append(app.connList, conn)
}

func (app *NetAssistantApp) updateStatus(msg string) {
	app.labelStatus.SetMarkup(msg)
}

func (app *NetAssistantApp) updateAllStatus(msg, curIP, curPort string) {
	app.labelStatus.SetMarkup(msg)
	app.entryCurAddr.SetText(curIP)
	app.entryCurPort.SetText(curPort)
}

func (app *NetAssistantApp) createConnect(serverType int, strIP, strPort string) error {
	addr := strIP + ":" + strPort
	if serverType == 0 { // TCP Client
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			go app.handler(conn)
			app.addConnection(conn)
			locallConnInfo := strings.Split(conn.LocalAddr().String(), ":")
			app.updateAllStatus("TCP client connection succeeds", locallConnInfo[0], locallConnInfo[1])
		} else {
			app.updateAllStatus(err.Error(), "", "")
			log.Error(err.Error())
			return err
		}
	}
	if serverType == 1 { // TCP Server
		listen, err := net.Listen("tcp", addr)
		if err == nil {

			app.updateStatus("TCP server connection succeeds")
			go func() {
				for {
					conn, err := listen.Accept()
					if err != nil {
						log.Error("accept err:", err)
						return
					}
					ss := conn.RemoteAddr().String()
					tips := fmt.Sprintf(`<span foreground="green">new connection:%s </span>`, ss)
					glib.IdleAdd(func() {
						app.labelStatus.SetMarkup(tips)
					})
					app.connList = append(app.connList, conn)
					go app.handler(conn)
				}
			}()
			app.updateAllStatus("TCP server connection succeeds", strIP, strPort)
			app.listener = listen
		} else {
			app.updateStatus(err.Error())
			log.Error(err)
			return err
		}
	}

	if serverType == 2 { // UDP Client
		conn, err := net.Dial("udp4", addr)
		if err == nil {
			go app.handler(conn)
			app.addConnection(conn)
			localConnInfo := strings.Split(conn.LocalAddr().String(), ":")
			app.updateAllStatus("UDP client connection succeeds", localConnInfo[0], localConnInfo[1])
		} else {
			app.updateStatus(err.Error())
			log.Error(err)
			return err
		}
	}

	if serverType == 3 { // UDP Server
		address, err := net.ResolveUDPAddr("udp4", addr)
		if err != nil {
			app.updateStatus("UDP server connection succeeds" + err.Error())
		} else {
			udpConn, err := net.ListenUDP("udp4", address)
			if err == nil {
				go app.handler(udpConn)
				app.addConnection(udpConn)
				localConnInfo := strings.Split(udpConn.LocalAddr().String(), ":")
				app.updateAllStatus("UDP server connection succeeds", localConnInfo[0], localConnInfo[1])
				app.labelLocalAddr.SetLabel("Taget UDP IP")
				app.labelLocalPort.SetLabel("Taget UDP Port")
				app.entryCurAddr.SetEditable(true)
				app.entryCurAddr.SetText("")
				app.entryCurPort.SetEditable(true)
				app.entryCurPort.SetText("")
			} else {
				app.updateStatus(err.Error())
				log.Error(err)
				return err
			}
		}
	}

	return nil
}

func (app *NetAssistantApp) disconnect(serverType int) error {
	if serverType == 1 {
		if app.listener != nil {
			app.listener.Close()
		}
	}

	for _, conn := range app.connList {
		conn.Close()
	}

	if serverType == 3 {
		app.labelLocalAddr.SetLabel(getI18nText(IT_LOCAL_IP))
		app.labelLocalPort.SetLabel(getI18nText(IT_LOCAL_PORT))
		app.entryCurAddr.SetEditable(false)
		app.entryCurAddr.SetText("")
		app.entryCurPort.SetEditable(false)
		app.entryCurPort.SetText("")
	}

	app.updateStatus(getI18nText(IT_WAIT_CONN))
	app.connList = []net.Conn{}
	return nil
}

func (app *NetAssistantApp) onBtnConnect(button *gtk.Button) {
	strIP, _ := app.entryIP.GetText()
	strPort, _ := app.entryPort.GetText()
	serverType := app.combProtoType.GetActive()
	label, _ := app.btnConnect.GetLabel()

	isDisconnect := label == getI18nText(IT_DISCONNECT)
	if isDisconnect {
		if err := app.disconnect(serverType); err == nil {
			app.btnConnect.SetLabel(getI18nText(IT_CONNECT))
			app.combProtoType.SetSensitive(true)
		}
	} else {
		if err := app.createConnect(serverType, strIP, strPort); err == nil {
			app.btnConnect.SetLabel(getI18nText(IT_DISCONNECT))
			app.combProtoType.SetSensitive(false)
		}
	}
}

func (app *NetAssistantApp) onBtnSend() {
	buff, err := app.tvDataSend.GetBuffer()
	if err != nil {
		log.Error(err)
	}

	start, end := buff.GetBounds()
	data, _ := buff.GetText(start, end, true)

	if app.cbAppendNewLine.GetActive() {
		data += "\r\n"
	}

	sendData := []byte(data)

	if app.cbSendByHex.GetActive() {
		data = strings.Replace(data, " ", "", -1)
		data = strings.Replace(data, "\n", "", -1)
		hexData, err := hex.DecodeString(data)
		if err != nil {
			log.Error(err)
			strTips := fmt.Sprintf(`<span foreground="red">😱%s</span>`, err)
			app.labelStatus.SetMarkup(strTips)
		} else {
			sendData = hexData
		}
		log.Info(hexData)
	}

	label, err := app.btnSend.GetLabel()
	if label != getI18nText(IT_SEND) {
		app.chanClose <- true
		app.btnSend.SetLabel(getI18nText(IT_SEND))
		return
	}

	if app.cbDataSourceCycleSend.GetActive() { // loop send
		app.btnSend.SetLabel(getI18nText(IT_STOP))
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
				case <-app.chanClose: // waiting close
					break END
				default:
					for _, conn := range app.connList { // range current connection
						if udpConnection, ok := conn.(*net.UDPConn); ok && app.combProtoType.GetActive() == 3 { // if the connection is UDP
							strIP, _ := app.entryCurAddr.GetText()
							strPort, _ := app.entryCurPort.GetText()
							address, err := net.ResolveUDPAddr("udp4", strIP+":"+strPort)
							if err == nil {
								udpConnection.WriteToUDP(sendData, address) // send udp message
							} else {
								log.Error(err)
							}

						} else {
							conn.Write(sendData) // send message
						}
					}
					if len(app.connList) == 0 {
						glib.IdleAdd(func() {
							app.labelStatus.SetText(getI18nText(IT_NO_CONN))
							app.btnSend.SetLabel(getI18nText(IT_SEND))
						})
						break END
					}
				}
				time.Sleep(time.Duration(cycleTime) * time.Millisecond)
			}
		}(cycle)
	} else { // once send
		for _, conn := range app.connList {
			if cc, ok := conn.(*net.UDPConn); ok && app.combProtoType.GetActive() == 3 {
				strIP, _ := app.entryCurAddr.GetText()
				strPort, _ := app.entryCurPort.GetText()
				address, err := net.ResolveUDPAddr("udp4", strIP+":"+strPort)
				if err == nil {
					cc.WriteToUDP(sendData, address)
				}
			} else {
				conn.Write(sendData)
			}
			log.Info("Write data", data)
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
	loader, _ := gdk.PixbufLoaderNew()
	data, _ := base64.StdEncoding.DecodeString(icon)
	loader.Write(data)
	buf, _ := loader.GetPixbuf()
	app.appWindow.SetIcon(buf)

	app.appWindow.SetBorderWidth(10)
	app.appWindow.SetTitle(getI18nText(IT_APP_NAME))

	// container
	windowContainer, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	windowContainerMiddle, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	windowContainerLeft, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	windowContainerRight, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	windowContainerBottom, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)

	// Left box
	// type of service frame
	frame, _ := gtk.FrameNew(getI18nText(IT_SETTING))
	frame.SetLabelAlign(0.1, 0.5)
	verticalBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10) //
	frame.Add(verticalBox)
	labelProtType, _ := gtk.LabelNew(getI18nText(IT_TYPE))
	labelProtType.SetXAlign(0)
	app.combProtoType, _ = gtk.ComboBoxTextNew()
	app.combProtoType.AppendText(getI18nText(IT_TCP_CLIENT))
	app.combProtoType.AppendText(getI18nText(IT_TCP_SERVER))
	app.combProtoType.AppendText(getI18nText(IT_UDP_CLIENT))
	app.combProtoType.AppendText(getI18nText(IT_UDP_SERVER))
	app.combProtoType.SetActive(0)
	verticalBox.PackStart(labelProtType, false, false, 0)
	verticalBox.PackStart(app.combProtoType, false, false, 0)
	labelIP, _ := gtk.LabelNew("IP")
	labelIP.SetXAlign(0)
	app.entryIP, _ = gtk.EntryNew()
	app.entryIP.SetText("127.0.0.1")
	verticalBox.PackStart(labelIP, false, false, 0)
	verticalBox.PackStart(app.entryIP, false, false, 0)
	labelPort, _ := gtk.LabelNew(getI18nText(IT_PORT))
	labelPort.SetXAlign(0)
	app.entryPort, _ = gtk.EntryNew()
	app.entryPort.SetText("8003")
	verticalBox.PackStart(labelPort, false, false, 0)
	verticalBox.PackStart(app.entryPort, false, false, 0)
	app.btnConnect, _ = gtk.ButtonNewWithLabel(getI18nText(IT_CONNECT))
	app.btnConnect.Connect("clicked", app.onBtnConnect)
	verticalBox.PackStart(app.btnConnect, false, false, 0)

	// Recv Settings, Send Settings
	notebookTab, _ := gtk.NotebookNew()
	label1, _ := gtk.LabelNew(getI18nText(IT_RECV_SETTINGS))
	label2, _ := gtk.LabelNew(getI18nText(IT_SEND_SETTINGS))

	//  Recv Settings
	frame1ContentBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	app.cbReceive2File, _ = gtk.CheckButtonNewWithLabel(getI18nText(IT_SAVE_TO_FILE))
	app.cbReceive2File.Connect("toggled", app.onCbReceive2File)
	app.cbDisplayDate, _ = gtk.CheckButtonNewWithLabel(getI18nText(IT_SHOW_RECV_TIME))
	app.cbHexDisplay, _ = gtk.CheckButtonNewWithLabel(getI18nText(IT_SHOW_HEX))
	app.cbPauseDisplay, _ = gtk.CheckButtonNewWithLabel(getI18nText(IT_PAUSE))
	btnHboxContainer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	app.btnSaveData, _ = gtk.ButtonNewWithLabel(getI18nText(IT_SAVE))
	app.btnSaveData.Connect("clicked", app.onBtnSaveData)
	app.btnClearRecvDisplay, _ = gtk.ButtonNewWithLabel(getI18nText(IT_CLEAR))
	app.btnClearRecvDisplay.Connect("clicked", app.onBtnClearRecvDisplay)

	btnHboxContainer.PackStart(app.btnSaveData, true, false, 0)
	btnHboxContainer.PackStart(app.btnClearRecvDisplay, true, false, 0)
	frame1ContentBox.PackStart(app.cbReceive2File, false, false, 0)
	frame1ContentBox.PackStart(app.cbDisplayDate, false, false, 0)
	frame1ContentBox.PackStart(app.cbHexDisplay, false, false, 0)
	frame1ContentBox.PackStart(app.cbPauseDisplay, false, false, 0)
	frame1ContentBox.PackStart(btnHboxContainer, false, false, 0)
	frame1ContentBox.SetBorderWidth(10)

	// Send Settings
	frame2ContentBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	app.cbAppendNewLine, _ = gtk.CheckButtonNewWithLabel(getI18nText(IT_APPEND_RN))
	app.cbAutoCleanAfterSend, _ = gtk.CheckButtonNewWithLabel(getI18nText(IT_AUTO_CLEAR))
	app.cbSendByHex, _ = gtk.CheckButtonNewWithLabel(getI18nText(IT_SEND_HEX))
	app.cbDataSourceCycleSend, _ = gtk.CheckButtonNewWithLabel(getI18nText(IT_SEND_CIRC))
	app.entryCycleTime, _ = gtk.EntryNew()
	app.entryCycleTime.SetPlaceholderText("default 1000(ms)")
	btnHboxContainer2, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	app.btnLoadData, _ = gtk.ButtonNewWithLabel(getI18nText(IT_LOAD_DATA))
	app.btnClearSendDisplay, _ = gtk.ButtonNewWithLabel(getI18nText(IT_CLEAR))
	app.btnLoadData.Connect("clicked", app.onBtnLoadData)
	app.btnClearSendDisplay.Connect("clicked", func() {
		buff, _ := app.tvDataSend.GetBuffer()
		buff.SetText("")
	})

	frame2ContentBox.PackStart(app.cbAppendNewLine, false, false, 0)
	frame2ContentBox.PackStart(app.cbAutoCleanAfterSend, false, false, 0)
	frame2ContentBox.PackStart(app.cbSendByHex, false, false, 0)
	frame2ContentBox.PackStart(app.cbDataSourceCycleSend, false, false, 0)
	frame2ContentBox.PackStart(app.entryCycleTime, false, false, 0)
	btnHboxContainer2.PackStart(app.btnLoadData, true, false, 0)
	btnHboxContainer2.PackStart(app.btnClearSendDisplay, true, false, 0)
	frame2ContentBox.PackStart(btnHboxContainer2, false, false, 0)
	frame2ContentBox.SetBorderWidth(10)

	frame1, _ := gtk.FrameNew("")
	frame1.Add(frame1ContentBox)
	frame2, _ := gtk.FrameNew("")
	frame2.Add(frame2ContentBox)

	notebookTab.AppendPage(frame1, label1)
	notebookTab.AppendPage(frame2, label2)

	// Data Received
	titleDataReceiveArea, _ := gtk.LabelNew(getI18nText(IT_DATA_RECVED))
	titleDataReceiveArea.SetXAlign(0)
	windowContainerRight.PackStart(titleDataReceiveArea, false, false, 0)
	app.swDataRec, _ = gtk.ScrolledWindowNew(nil, nil)
	app.tvDataReceive, _ = gtk.TextViewNew()
	app.tvDataReceive.SetEditable(false)
	app.tvDataReceive.SetWrapMode(gtk.WRAP_CHAR)
	app.swDataRec.Add(app.tvDataReceive)
	windowContainerRight.PackStart(app.swDataRec, true, true, 0)

	// Local info
	middleContainer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	app.labelLocalAddr, _ = gtk.LabelNew(getI18nText(IT_LOCAL_IP))
	app.entryCurAddr, _ = gtk.EntryNew()
	app.entryCurAddr.SetEditable(false)
	app.labelLocalPort, _ = gtk.LabelNew(getI18nText(IT_LOCAL_PORT))
	app.entryCurPort, _ = gtk.EntryNew()
	app.entryCurPort.SetEditable(false)
	middleContainer.PackStart(app.labelLocalAddr, false, false, 0)
	middleContainer.PackStart(app.entryCurAddr, false, false, 0)
	middleContainer.PackStart(app.labelLocalPort, false, false, 0)
	middleContainer.PackStart(app.entryCurPort, false, false, 0)
	windowContainerRight.PackStart(middleContainer, false, false, 0)

	// send area
	bottomContainer, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	scrollerDataSend, _ := gtk.ScrolledWindowNew(nil, nil)
	app.tvDataSend, _ = gtk.TextViewNew()
	app.tvDataSend.SetWrapMode(gtk.WRAP_CHAR)
	scrollerDataSend.Add(app.tvDataSend)
	scrollerDataSend.SetSizeRequest(-1, 180)
	boxSendBtn, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	app.btnSend, _ = gtk.ButtonNewWithLabel(getI18nText(IT_SEND))
	app.btnSend.Connect("clicked", app.onBtnSend)
	boxSendBtn.PackEnd(app.btnSend, false, false, 0)
	app.btnSend.SetSizeRequest(80, -1)
	bottomContainer.PackStart(scrollerDataSend, true, true, 0)
	bottomContainer.PackEnd(boxSendBtn, false, false, 0)
	windowContainerRight.PackStart(bottomContainer, false, false, 0)

	// bottom area
	app.labelStatus, _ = gtk.LabelNew("")
	app.labelStatus.SetMarkup(getI18nText(IT_WAIT_CONN))
	windowContainerBottom.PackStart(app.labelStatus, true, false, 0)
	app.labelSendCount, _ = gtk.LabelNew(getI18nText(IT_SEND_COUNT))
	windowContainerBottom.PackStart(app.labelSendCount, true, false, 0)
	app.labelReceveCount, _ = gtk.LabelNew(getI18nText(IT_RECEVER_COUNT))
	windowContainerBottom.PackStart(app.labelReceveCount, true, false, 0)
	app.btnCleanCount, _ = gtk.ButtonNewWithLabel(getI18nText(IT_RESET))
	app.btnCleanCount.Connect("clicked", app.onBtnCleanCount)

	windowContainerBottom.PackEnd(app.btnCleanCount, false, false, 0)

	app.appWindow.Add(windowContainer)

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

func init() {
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfile} func:%{longfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)

	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend1Leveled := logging.AddModuleLevel(backend1)
	formatter := logging.NewBackendFormatter(backend1, format)
	backend1Leveled = logging.SetBackend(formatter)
	if os.Getenv("NET_ASS_DEBUG") == "on" {
		backend1Leveled.SetLevel(logging.DEBUG, "")
		fmt.Println("NET_ASS_DEBUG", "on")
	} else {
		backend1Leveled.SetLevel(logging.CRITICAL, "")
	}
}

func main() {
	const appID = "com.github.baloneo.netassistant"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_NON_UNIQUE)

	if err != nil {
		log.Fatal("Could not create application.", err)
	}

	app := NetAssistantAppNew()
	application.Connect("activate", app.doActivate)

	application.Run(os.Args)
}
