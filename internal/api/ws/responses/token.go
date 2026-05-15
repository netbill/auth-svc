package responses

import "github.com/gorilla/websocket"

func QRToken(conn *websocket.Conn, token string) error {
	return conn.WriteJSON(struct {
		QRToken string `json:"qr_token"`
	}{QRToken: token})
}

func AuthTokensPair(conn *websocket.Conn, payload []byte) error {
	return conn.WriteMessage(websocket.TextMessage, payload)
}

func TokensExpiredError(conn *websocket.Conn) error {
	return conn.WriteJSON(struct {
		Error string `json:"error"`
	}{Error: "tokens expired"})
}
