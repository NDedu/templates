package main

import (
	"net/http"
	"sync"
	"time"
	"webProj/internal/models"

	"github.com/gorilla/websocket"
)

const (
	wsWriteWait  = 10 * time.Second
	wsPongWait   = 60 * time.Second
	wsPingPeriod = (wsPongWait * 9) / 10
	wsMaxMsgSize = 4096
)

type WebSocketConnection struct {
	*websocket.Conn
	mu        sync.Mutex
	userUUID  string
	isAdmin   bool
	done      chan struct{}
	closeOnce sync.Once
}

type WsPayload struct {
	Action   string `json:"action"`
	Message  string `json:"message"`
	UserName string `json:"username"`
	UserUUID string `json:"user_uuid"`
	sender   *WebSocketConnection
}

type WsJSONResponse struct {
	Action   string `json:"action"`
	Message  string `json:"message"`
	UserUUID string `json:"user_uuid"`
}

type wsHub struct {
	mu      sync.Mutex
	clients map[*WebSocketConnection]bool
}

func (c *WebSocketConnection) safeWriteJSON(v any) error {

	c.mu.Lock()
	defer c.mu.Unlock()
	c.SetWriteDeadline(time.Now().Add(wsWriteWait))

	return c.WriteJSON(v)
}

func (c *WebSocketConnection) safeWritePing() error {

	c.mu.Lock()
	defer c.mu.Unlock()
	c.SetWriteDeadline(time.Now().Add(wsWriteWait))

	return c.WriteMessage(websocket.PingMessage, nil)
}

func (c *WebSocketConnection) safeWriteClose(msg string) {

	c.mu.Lock()
	defer c.mu.Unlock()
	c.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseGoingAway, msg),
		time.Now().Add(wsWriteWait),
	)
}

func (app *application) WsEndpoint(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)
	if user == nil {

		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ws, err := app.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {

		app.errorLog.Println(err)
		return
	}

	app.infoLog.Printf("WebSocket client connected: %s (user %s)", r.RemoteAddr, user.UUID)

	conn := &WebSocketConnection{
		Conn:     ws,
		userUUID: user.UUID,
		isAdmin:  user.Role == models.RoleAdmin,
		done:     make(chan struct{}),
	}

	ws.SetReadLimit(wsMaxMsgSize)
	ws.SetReadDeadline(time.Now().Add(wsPongWait))
	ws.SetPongHandler(func(string) error {

		ws.SetReadDeadline(time.Now().Add(wsPongWait))
		return nil
	})

	app.wsHub.mu.Lock()
	app.wsHub.clients[conn] = true
	app.wsHub.mu.Unlock()

	if err := conn.safeWriteJSON(WsJSONResponse{Message: "Connected to server"}); err != nil {

		app.errorLog.Println(err)
		app.removeWsClient(conn)
		return
	}

	go app.wsPingPump(conn)
	go app.wsReadPump(conn)
}

func (app *application) wsReadPump(conn *WebSocketConnection) {

	defer app.removeWsClient(conn)

	for {

		var payload WsPayload
		if err := conn.ReadJSON(&payload); err != nil {

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {

				app.errorLog.Printf("WebSocket read error: %v", err)
			}
			return
		}
		payload.sender = conn
		select {
		case app.wsChan <- payload:
		default:
			app.errorLog.Printf("WebSocket channel full, dropping message from %s", conn.userUUID)
		}
	}
}

func (app *application) wsPingPump(conn *WebSocketConnection) {

	ticker := time.NewTicker(wsPingPeriod)
	defer ticker.Stop()

	for {

		select {

		case <-ticker.C:
			if err := conn.safeWritePing(); err != nil {
				return
			}
		case <-conn.done:
			return
		case <-app.ctx.Done():
			conn.safeWriteClose("server shutting down")
			return
		}
	}
}

func (app *application) ListenToWsChannel() {

	for {

		select {

		case event := <-app.wsChan:
			switch event.Action {
			case "deleteUser":
				if event.sender == nil || !event.sender.isAdmin {

					break
				}
				app.sendToUser(event.UserUUID, WsJSONResponse{
					Action:   "logout",
					Message:  "Your account has been deleted",
					UserUUID: event.UserUUID,
				})
			}
		case <-app.ctx.Done():
			app.closeAllWsClients()
			// drain any remaining buffered payloads
			for {

				select {
				case <-app.wsChan:
				default:
					return
				}
			}
		}
	}
}

func (app *application) sendToUser(userUUID string, response WsJSONResponse) {

	app.wsHub.mu.Lock()
	var targetClients []*WebSocketConnection

	for client := range app.wsHub.clients {

		if client.userUUID == userUUID {

			targetClients = append(targetClients, client)
		}
	}
	app.wsHub.mu.Unlock()

	for _, client := range targetClients {

		if err := client.safeWriteJSON(response); err != nil {

			app.errorLog.Printf("WebSocket write error for user %s: %v", userUUID, err)
			app.removeWsClient(client)
		}
	}
}

func (app *application) broadcastToAll(response WsJSONResponse) {

	app.wsHub.mu.Lock()
	var targetClients []*WebSocketConnection
	for client := range app.wsHub.clients {

		targetClients = append(targetClients, client)
	}
	app.wsHub.mu.Unlock()

	for _, client := range targetClients {

		if err := client.safeWriteJSON(response); err != nil {

			app.errorLog.Printf("WebSocket write error on %s: %v", response.Action, err)
			app.removeWsClient(client)
		}
	}
}

func (app *application) removeWsClient(conn *WebSocketConnection) {

	app.wsHub.mu.Lock()
	_, ok := app.wsHub.clients[conn]
	if ok {

		delete(app.wsHub.clients, conn)
	}
	app.wsHub.mu.Unlock()

	if ok {

		conn.closeOnce.Do(func() {
			close(conn.done)
			conn.Close()
		})
	}
}

func (app *application) closeAllWsClients() {

	app.wsHub.mu.Lock()
	clients := make([]*WebSocketConnection, 0, len(app.wsHub.clients))
	for client := range app.wsHub.clients {

		clients = append(clients, client)
		delete(app.wsHub.clients, client)
	}
	app.wsHub.mu.Unlock()

	for _, client := range clients {

		client.safeWriteClose("server shutting down")
		client.closeOnce.Do(func() {
			close(client.done)
			client.Close()
		})
	}
}
