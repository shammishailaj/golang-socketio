package transport

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mtfelian/golang-socketio/logging"
)

const (
	upgradeFailed = "Upgrade failed: "

	WsDefaultPingInterval   = 30 * time.Second
	WsDefaultPingTimeout    = 60 * time.Second
	WsDefaultReceiveTimeout = 60 * time.Second
	WsDefaultSendTimeout    = 60 * time.Second
	WsDefaultBufferSize     = 1024 * 32
)

// WebsocketTransportParams is a parameters for getting non-default websocket transport
type WebsocketTransportParams struct {
	Headers http.Header
}

var (
	ErrorBinaryMessage     = errors.New("Binary messages are not supported")
	ErrorBadBuffer         = errors.New("Buffer error")
	ErrorPacketWrong       = errors.New("Wrong packet type error")
	ErrorMethodNotAllowed  = errors.New("Method not allowed")
	ErrorHttpUpgradeFailed = errors.New("Http upgrade failed")
)

type WebsocketConnection struct {
	socket    *websocket.Conn
	transport *WebsocketTransport
}

func (wsc *WebsocketConnection) GetMessage() (message string, err error) {
	logging.Log().Debug("GetMessage ws begin")
	wsc.socket.SetReadDeadline(time.Now().Add(wsc.transport.ReceiveTimeout))

	msgType, reader, err := wsc.socket.NextReader()
	if err != nil {
		logging.Log().Debug("ws reading err ", err)
		return "", err
	}

	//support only text messages exchange
	if msgType != websocket.TextMessage {
		logging.Log().Debug("ws reading err ErrorBinaryMessage")
		return "", ErrorBinaryMessage
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		logging.Log().Debug("ws reading err ErrorBadBuffer")
		return "", ErrorBadBuffer
	}

	text := string(data)
	logging.Log().Debug("GetMessage ws text ", text)

	//empty messages are not allowed
	if len(text) == 0 {
		logging.Log().Debug("ws reading err ErrorPacketWrong")
		return "", ErrorPacketWrong
	}

	return text, nil
}

func (wsc *WebsocketTransport) SetSid(sid string, conn Connection) {}

func (wsc *WebsocketConnection) WriteMessage(message string) error {
	logging.Log().Debug("WriteMessage ws ", message)
	wsc.socket.SetWriteDeadline(time.Now().Add(wsc.transport.SendTimeout))

	writer, err := wsc.socket.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	if _, err := writer.Write([]byte(message)); err != nil {
		return err
	}

	return writer.Close()
}

func (wsc *WebsocketConnection) Close() {
	logging.Log().Debug("ws close")
	wsc.socket.Close()
}

func (wsc *WebsocketConnection) PingParams() (time.Duration, time.Duration) {
	return wsc.transport.PingInterval, wsc.transport.PingTimeout
}

type WebsocketTransport struct {
	PingInterval   time.Duration
	PingTimeout    time.Duration
	ReceiveTimeout time.Duration
	SendTimeout    time.Duration

	BufferSize int

	Headers http.Header
}

func (wst *WebsocketTransport) Connect(url string) (Connection, error) {
	dialer := websocket.Dialer{}
	socket, _, err := dialer.Dial(url, wst.Headers)
	if err != nil {
		return nil, err
	}

	return &WebsocketConnection{socket, wst}, nil
}

func (wst *WebsocketTransport) HandleConnection(w http.ResponseWriter, r *http.Request) (Connection, error) {

	if r.Method != http.MethodGet {
		http.Error(w, upgradeFailed+ErrorMethodNotAllowed.Error(), 503)
		return nil, ErrorMethodNotAllowed
	}

	socket, err := websocket.Upgrade(w, r, nil, wst.BufferSize, wst.BufferSize)
	if err != nil {
		http.Error(w, upgradeFailed+err.Error(), 503)
		return nil, ErrorHttpUpgradeFailed
	}

	return &WebsocketConnection{socket, wst}, nil
}

// Websocket connection do not require any additional processing
func (wst *WebsocketTransport) Serve(w http.ResponseWriter, r *http.Request) {}

// Returns websocket connection with default params
func GetDefaultWebsocketTransport() *WebsocketTransport {
	return &WebsocketTransport{
		PingInterval:   WsDefaultPingInterval,
		PingTimeout:    WsDefaultPingTimeout,
		ReceiveTimeout: WsDefaultReceiveTimeout,
		SendTimeout:    WsDefaultSendTimeout,
		BufferSize:     WsDefaultBufferSize,

		Headers: nil,
	}
}

// GetWebsocketTransport returns websocket transport with additional params
func GetWebsocketTransport(params WebsocketTransportParams) *WebsocketTransport {
	tr := GetDefaultWebsocketTransport()
	tr.Headers = params.Headers
	return tr
}
