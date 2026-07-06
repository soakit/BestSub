package service

import (
	"sync"
	"time"
)

const maxEvents = 1000

// RefreshEvent 表示刷新过程中的一条事件。
type RefreshEvent struct {
	SubID   string         `json:"sub_id"`
	Type    string         `json:"type"` // "start", "progress", "done", "error"
	Payload map[string]any `json:"payload"`
}

var (
	hubMu     sync.Mutex
	hubEvents []RefreshEvent
	hubStart  int
	hubNotify = make(chan struct{}, 1)
	hubSend   int
)

// Subscribe 注册当前连接为订阅者。发送端始终从缓冲区当前位置开始读取。
// 阻塞直到连接断开（ctx 取消）。
func Subscribe(send func(RefreshEvent) error, ctxDone <-chan struct{}) {
	go func() {
		for {
			hubMu.Lock()
			if hubSend < hubStart {
				hubSend = hubStart
			}
			var batch []RefreshEvent
			for hubSend < hubStart+len(hubEvents) {
				batch = append(batch, hubEvents[hubSend-hubStart])
				hubSend++
			}
			hubMu.Unlock()

			for _, ev := range batch {
				select {
				case <-ctxDone:
					return
				default:
				}
				if err := send(ev); err != nil {
					return
				}
			}

			if len(batch) > 0 {
				continue
			}

			select {
			case <-ctxDone:
				return
			case <-hubNotify:
			case <-time.After(15 * time.Second):
				send(RefreshEvent{Type: "heartbeat"})
			}
		}
	}()
	<-ctxDone
}

// Emit 发布一条事件到环形缓冲区。
func Emit(subID, eventType string, payload map[string]any) {
	hubMu.Lock()
	hubEvents = append(hubEvents, RefreshEvent{
		SubID:   subID,
		Type:    eventType,
		Payload: payload,
	})
	if len(hubEvents) > maxEvents {
		overflow := len(hubEvents) - maxEvents
		hubStart += overflow
		hubEvents = hubEvents[overflow:]
	}
	hubMu.Unlock()

	select {
	case hubNotify <- struct{}{}:
	default:
	}
}
