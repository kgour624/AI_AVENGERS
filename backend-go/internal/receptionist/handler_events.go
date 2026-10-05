package receptionist

import (
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) StreamEvents(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		auth := c.GetHeader("Authorization")
		token = strings.TrimPrefix(auth, "Bearer ")
	}
	if token == "" {
		c.AbortWithStatusJSON(401, gin.H{"code": "UNAUTHORIZED", "msg": "token missing"})
		return
	}

	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"code": "INVALID_SESSION_ID"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	var events []ConversationEvent
	if h.store != nil {
		events, _ = h.store.ListEvents(c.Request.Context(), sessionID, 0, 100)
	}

	c.Stream(func(w io.Writer) bool {
		for _, e := range events {
			c.SSEvent(e.Type, string(e.Payload))
		}
		c.Writer.Flush()

		if h.publisher != nil {
			sub := h.publisher.Subscribe(c.Request.Context(), sessionIDStr)
			if sub != nil {
				defer sub.Close()
				ch := sub.Channel()
				ticker := time.NewTicker(15 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-c.Request.Context().Done():
						return false
					case msg, ok := <-ch:
						if !ok {
							return false
						}
						c.SSEvent("message", msg.Payload)
						c.Writer.Flush()
					case <-ticker.C:
						c.SSEvent("ping", "keepalive")
						c.Writer.Flush()
					}
				}
			}
		}
		return false
	})
}
