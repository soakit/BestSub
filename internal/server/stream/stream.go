package stream

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const DefaultMaxEvents = 100

type event struct {
	typ  string // typ 是 SSE 事件名。
	data any    // data 是写给前端的事件数据。
}

type Stream struct {
	mu     sync.Mutex              // mu 保护环形缓冲区、读写指针和连接信号。
	events [DefaultMaxEvents]event // events 是固定长度的环形缓冲区。
	read   int                     // read 是下一条要发送给前端的事件序号。
	write  int                     // write 是下一条要写入缓冲区的事件序号。
	notify chan struct{}           // notify 用于唤醒正在等待新事件的订阅连接。
	client chan struct{}           // client 是当前连接的退出信号，新连接进入时关闭旧信号以抛弃旧连接。
}

func New() *Stream {
	return &Stream{
		notify: make(chan struct{}),
	}
}

func (s *Stream) Emit(typ string, data any) {
	s.mu.Lock()
	s.events[s.write%DefaultMaxEvents] = event{typ: typ, data: data}
	s.write++
	// 写指针追上读指针时丢弃最旧未读事件，避免覆盖后读到错误槽位。
	if s.write-s.read > DefaultMaxEvents {
		s.read = s.write - DefaultMaxEvents
	}
	close(s.notify)
	s.notify = make(chan struct{})
	s.mu.Unlock()
}

func (s *Stream) Subscribe(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	done := make(chan struct{})
	s.mu.Lock()
	if s.client != nil {
		close(s.client)
	}
	s.client = done
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		if s.client == done {
			s.client = nil
		}
		s.mu.Unlock()
	}()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		s.mu.Lock()
		if s.read < s.write {
			seq := s.read
			event := s.events[seq%DefaultMaxEvents]
			s.mu.Unlock()
			select {
			case <-c.Request.Context().Done():
				return
			case <-done:
				return
			default:
			}
			c.SSEvent(event.typ, event.data)
			c.Writer.Flush()

			s.mu.Lock()
			// 旧连接被新连接替换后不能推进 read，否则新连接会跳过未发送事件。
			if s.client == done && s.read == seq {
				s.read++
			}
			s.mu.Unlock()
			continue
		}
		notify := s.notify
		s.mu.Unlock()

		select {
		case <-c.Request.Context().Done():
			return
		case <-done:
			return
		case <-notify:
		case <-ticker.C:
			if _, err := c.Writer.Write([]byte(": ping\n\n")); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}
