package mino

import (
	"io"
)

// 连接会话
type Session struct {
	// 本地连接
	loc io.ReadWriteCloser
	// 远程连接
	rmt io.ReadWriteCloser
}

// 创建会话
func NewSession(loc, rmt io.ReadWriteCloser) *Session {
	return &Session{
		loc: loc,
		rmt: rmt,
	}
}

// 传输数据
func (s *Session) Transport() (up, down int64) {
	var _closed = make(chan struct{})
	go func() {
		// send local -> remote
		up, _ = io.Copy(s.rmt, s.loc)
		_closed <- struct{}{}
	}()
	go func() {
		// send remote -> down
		down, _ = io.Copy(s.loc, s.rmt)
		_closed <- struct{}{}
	}()
	<-_closed
	_ = s.loc.Close()
	_ = s.rmt.Close()
	return
}
