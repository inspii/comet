package internal

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSSessionOption struct {
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	PingPongInterval time.Duration
}

type wsReadCmd struct {
	buf    []byte
	resN   int
	resErr error
}

type wsWriteCmd struct {
	data   []byte
	resN   int
	resErr error
}

type WSSession struct {
	ctx              context.Context
	conn             *websocket.Conn
	readTimeout      time.Duration
	writeTimeout     time.Duration
	pingPongInterval time.Duration

	once   sync.Once
	waiter sync.WaitGroup
	pings  int
	pongs  int

	readQueue  chan *wsReadCmd
	writeQueue chan *wsWriteCmd
}

func NewWSSession(ctx context.Context, conn *websocket.Conn, option *WSSessionOption) *WSSession {
	if option == nil {
		option = &WSSessionOption{
			ReadTimeout:      15 * time.Second,
			WriteTimeout:     15 * time.Second,
			PingPongInterval: 30 * time.Second,
		}
	}
	rw := &WSSession{
		ctx:              ctx,
		conn:             conn,
		readTimeout:      option.ReadTimeout,
		writeTimeout:     option.WriteTimeout,
		pingPongInterval: option.PingPongInterval,
		readQueue:        make(chan *wsReadCmd),
		writeQueue:       make(chan *wsWriteCmd),
	}
	rw.waiter.Add(1)

	go rw.readPump()
	go rw.writePump()
	return rw
}

func (s *WSSession) Read(p []byte) (int, error) {
	c := &wsReadCmd{buf: p}
	s.readQueue <- c
	return c.resN, c.resErr
}

func (s *WSSession) Write(data []byte) (int, error) {
	c := &wsWriteCmd{data: data}
	s.writeQueue <- c
	return c.resN, c.resErr
}

func (s *WSSession) Wait() {
	s.waiter.Wait()
}

// 防止并发读
func (s *WSSession) readPump() {
	defer s.once.Do(func() {
		s.waiter.Done()
	})

	if err := s.conn.SetReadDeadline(time.Now().Add(s.pingPongInterval)); err != nil {
		return
	}
	s.conn.SetPongHandler(func(string) error {
		// TODO
		return nil
	})

	for {
		select {
		case cmd := <-s.readQueue:
			cmd.resN, cmd.resErr = s.read(cmd.buf)
			if cmd.resErr != nil {
				return
			}
		case <-s.ctx.Done():
			return
		}
	}
}

// 防止并发发送数据或Ping
func (s *WSSession) writePump() {
	defer s.once.Do(func() {
		s.waiter.Done()
	})

	ticker := time.NewTicker(s.pingPongInterval)
	for {
		select {
		case cmd := <-s.writeQueue:
			if cmd.resN, cmd.resErr = s.write(cmd.data); cmd.resErr != nil {
				return
			}
		case <-ticker.C:
			deadline := time.Now().Add(s.writeTimeout)
			if err := s.conn.SetWriteDeadline(deadline); err != nil {
				return
			}
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

			s.pings++
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *WSSession) read(p []byte) (int, error) {
	_, reader, err := s.conn.NextReader()
	if err != nil {
		return 0, err
	}
	deadline := time.Now().Add(s.readTimeout)
	if err := s.conn.SetReadDeadline(deadline); err != nil {
		return 0, err
	}

	return reader.Read(p)
}

func (s *WSSession) write(data []byte) (int, error) {
	deadline := time.Now().Add(s.writeTimeout)
	if err := s.conn.SetWriteDeadline(deadline); err != nil {
		_ = s.conn.WriteMessage(websocket.CloseMessage, []byte{})
		return 0, err
	}

	writer, err := s.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}

	n, err := writer.Write(data)
	if err != nil {
		return n, err
	}

	err = writer.Close()
	return n, err
}
