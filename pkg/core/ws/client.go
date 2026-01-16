package ws

import (
	"context"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func Send(ctx context.Context, conn *websocket.Conn, msg Message) error {
	return wsjson.Write(ctx, conn, msg)
}

func Read(ctx context.Context, conn *websocket.Conn) (*Message, error) {
	var msg Message
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
