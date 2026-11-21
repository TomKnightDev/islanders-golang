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
	var pending *resources.ConnectRequestContents // a new name waiting for its class
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
			if req.Username == "" {
				sess.push(resources.NewFailedToConnectMessage("Enter a name to play."))
				return
			}
			// A new name picks a class before the character exists.
			known := make(chan bool, 1)
			world.Do(func(w *World) { known <- w.exists(req.Username) })
			if !<-known {
				pending = &req
				sess.push(resources.NewChooseClassMessage())
				continue
			}
			if p = loginOrFail(sess, req, ""); p == nil {
				return
			}

		case resources.CreateCharacterMessage:
			if p != nil || pending == nil {
				continue
			}
			var create resources.CreateCharacterContents
			if err := resources.DecodeContents(message.Contents, &create); err != nil {
				log.Println("Bad create character:", err)
				continue
			}
			if _, ok := resources.FindClass(create.Class); !ok {
				continue
			}
			if p = loginOrFail(sess, *pending, create.Class); p == nil {
				return
			}
			pending = nil

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

		case resources.ShopCommandMessage:
			if p == nil {
				continue
			}
			var cmd resources.ShopCommandContents
			if err := resources.DecodeContents(message.Contents, &cmd); err != nil {
				log.Println("Bad shop command:", err)
				continue
			}
			world.Do(func(w *World) { w.shopCommand(p, cmd) })

		case resources.EquipCommandMessage:
			if p == nil {
				continue
			}
			var cmd resources.EquipCommandContents
			if err := resources.DecodeContents(message.Contents, &cmd); err != nil {
				log.Println("Bad equip command:", err)
				continue
			}
			world.Do(func(w *World) { w.equipCommand(p, cmd) })

		case resources.UpdateMessage:
			// Older clients still send their own position. The server owns
			// movement now, so these are ignored.

		default:
			log.Printf("Message type: %s not handled", message.MessageType)
		}
	}
}

// loginOrFail logs a session in, first creating the character as class if
// that's given and the name is still free. On failure it tells the client
// and returns nil.
func loginOrFail(sess *session, req resources.ConnectRequestContents, class resources.ClassID) *player {
	reply := make(chan loginResult, 1)
	world.Do(func(w *World) {
		if c, ok := resources.FindClass(class); ok && !w.exists(req.Username) {
			w.createCharacter(req.Username, req.Password, c)
		}
		p, err := w.login(req.Username, req.Password, sess)
		reply <- loginResult{p, err}
	})
	res := <-reply
	if res.err != nil {
		sess.push(resources.NewFailedToConnectMessage(res.err.Error()))
		return nil
	}
	log.Printf("%s connected", req.Username)
	return res.p
}
