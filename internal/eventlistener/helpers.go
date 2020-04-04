package eventlistener

import (
	"github.com/gorilla/websocket"
	"time"
)

func pingMessage(conn *websocket.Conn, writeWait time.Duration) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	return conn.WriteMessage(websocket.PingMessage, nil)
}

func closeMessage(conn *websocket.Conn, writeWait time.Duration) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	return conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func writeMessage(conn *websocket.Conn, writeWait time.Duration, message []byte) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}

	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	_, errWrite := w.Write(message)
	err = w.Close()

	if errWrite != nil {
		return errWrite
	}

	return err
}
