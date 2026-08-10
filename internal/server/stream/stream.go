package stream

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type event struct {
	typ  string // typ 是 SSE 事件名。
	data any    // data 是写给前端的事件数据。
}

// Stream 向当前唯一的 SSE 连接实时发送事件。
type Stream struct {
	mu     sync.Mutex    // mu 保护当前连接的事件通道和退出信号。
	events chan event    // events 是当前连接专属的实时事件通道，无连接时为 nil。
	client chan struct{} // client 是当前连接的退出信号，新连接进入时关闭旧信号。
}

// New 创建实时事件流。
func New() *Stream {
	return &Stream{}
}

// Emit 向当前连接发送事件，无连接时直接丢弃。
func (s *Stream) Emit(typ string, data any) {
	s.mu.Lock()
	events := s.events
	client := s.client
	s.mu.Unlock()
	if events == nil {
		return
	}
	select {
	case events <- event{typ: typ, data: data}:
	case <-client:
	}
}

// Subscribe 建立新的 SSE 连接并关闭已有连接。
func (s *Stream) Subscribe(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	events := make(chan event)
	done := make(chan struct{})
	s.mu.Lock()
	if s.client != nil {
		close(s.client)
	}
	s.events = events
	s.client = done
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		if s.client == done {
			s.events = nil
			s.client = nil
			close(done)
		}
		s.mu.Unlock()
	}()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-done:
			return
		case event := <-events:
			// 事件和退出信号同时就绪时优先终止旧连接，避免替换后继续写入。
			select {
			case <-c.Request.Context().Done():
				return
			case <-done:
				return
			default:
			}
			c.SSEvent(event.typ, event.data)
			c.Writer.Flush()
		case <-ticker.C:
			if _, err := c.Writer.Write([]byte(": ping\n\n")); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}
