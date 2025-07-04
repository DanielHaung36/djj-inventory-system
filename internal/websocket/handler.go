package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// ServeWS upgrades HTTP => WS and registers conn under :topic
func ServeWS(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		topic := c.Param("topic") // e.g. "customers", "quotes"
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		hub.Register(topic, conn)
		defer hub.Unregister(topic, conn)
		// keep connection alive until client disconnects
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}
}
