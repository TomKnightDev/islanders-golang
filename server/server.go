package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/tomknightdev/islanders-golang/resources"
)

var (
	world    = NewWorld(GenerateWorldMap())
	upgrader = websocket.Upgrader{} // use default options
)

type loginResult struct {
	p   *player
	err error
}

// connect reads one client's messages and turns them into world commands.
// It never touches game state itself.
func connect(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Upgrade error:", err)
		return
	}
	sess := newSession(conn)

	var p *player
	defer func() {
		if p != nil {
			world.Do(func(w *World) { w.disconnect(p, sess) })
		} else {
			sess.close()
		}
	}()

	for {
		message := &resources.Message{}
		if err := conn.ReadJSON(message); err != nil {
			log.Println("Connect read error:", err)
			return
		}

		switch message.MessageType {
		case resources.ConnectRequestMessage:
			if p != nil {
				continue
			}
			var req resources.ConnectRequestContents
			if err := resources.DecodeContents(message.Contents, &req); err != nil {
				log.Println("Bad connect request:", err)
				return
			}
			reply := make(chan loginResult, 1)
			world.Do(func(w *World) {
				p, err := w.login(req.Username, req.Password, sess)
				reply <- loginResult{p, err}
			})
			res := <-reply
			if res.err != nil {
				sess.push(resources.NewFailedToConnectMessage(res.err.Error()))
				return
			}
			p = res.p
			log.Printf("%s connected", req.Username)

		case resources.ChatMessage:
			text, ok := message.Contents.(string)
			if p == nil || !ok {
				continue
			}
			world.Do(func(w *World) { w.chat(p, text) })

		case resources.MoveRequestMessage:
			if p == nil {
				continue
			}
			var req resources.MoveRequestContents
			if err := resources.DecodeContents(message.Contents, &req); err != nil {
				log.Println("Bad move request:", err)
				continue
			}
			world.Do(func(w *World) { w.click(p, tile{req.X, req.Y}) })

		case resources.ActionCommandMessage:
			if p == nil {
				continue
			}
			var cmd resources.ActionCommandContents
			if err := resources.DecodeContents(message.Contents, &cmd); err != nil {
				log.Println("Bad action command:", err)
				continue
			}
			world.Do(func(w *World) { w.actionCommand(p, cmd) })

		case resources.UpdateMessage:
			// Older clients still send their own position. The server owns
			// movement now, so these are ignored.

		default:
			log.Printf("Message type: %s not handled", message.MessageType)
		}
	}
}
