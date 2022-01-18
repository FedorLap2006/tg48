package mk48

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"nhooyr.io/websocket"
)

type API struct {
	ws       *websocket.Conn
	session  *websocket.Conn
	Handlers struct {
		Message           func(interface{})
		LeaderboardUpdate func(LeaderboardUpdate)
	}
}

func send(conn *websocket.Conn, data Message) error {
	m, err := json.Marshal(map[string]interface{}{
		data.Name(): data,
	})
	if err != nil {
		return err
	}

	return conn.Write(context.TODO(), websocket.MessageText, m)
}

func recv(conn *websocket.Conn) (e *Event, err error) {
	var raw []byte
	_, raw, err = conn.Read(context.Background())
	if err != nil {
		return
	}

	err = json.Unmarshal(raw, &e)
	return
}

const recvTimeout = time.Millisecond * 300

func (api *API) Listen() {
	for {
		e, err := recv(api.ws)
		if err != nil {
			log.Panic(err)
		}

		switch {
		case e.LeaderboardUpdate != nil:
			if api.Handlers.LeaderboardUpdate != nil {
				api.Handlers.LeaderboardUpdate(*e.LeaderboardUpdate)
			}
		}
	}
}

func (api *API) Close() error {
	if api.session != nil {
		err := api.session.Close(websocket.StatusGoingAway, "")
		if err != nil {
			return err
		}
	}
	return api.ws.Close(websocket.StatusGoingAway, "")
}

func New() (api *API, err error) {
	ws, _, err := websocket.Dial(context.Background(), EndpointWebsocket+"/?format=json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the websocket: %w", err)
	}
	err = send(ws, CreateSession{
		GameID: "Mk48",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create a session: %w", err)
	}

	return &API{ws: ws}, nil
}
