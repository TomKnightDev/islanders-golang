# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Islanders is a small multiplayer top-down 2D game. One Go module (`github.com/tomknightdev/islanders-golang`) holds three packages:

- `server/`: a websocket game server (gorilla/websocket). It owns world generation, the player registry and the NPCs.
- `client/`: a desktop client built on Ebiten v2, with imgui for menus and chat. It needs **cgo/GCC** to build.
- `resources/`: shared message types, tile constants and the `WorldMap` struct. Both sides import it.

## Commands

```bash
make build          # builds bin/server and bin/client
make run-server     # go run ./server/ -addr 0.0.0.0:8285 (override with SERVER_ADDR=...)
make run-client     # go run ./client/
go build ./server/  # what CI runs; the client is not built in CI
go vet ./...
go test ./...       # there are no real tests yet; client/client_test.go is commented out
go test ./client -run TestName   # single test
```

Without `-addr`, the server binds to `<outbound IP>:8285`. The client's main menu defaults to the production server (`134.209.26.73:8285`). Change the address to `localhost:8285` to test locally.

## Deployment

Every push to `master` triggers `.github/workflows/server.yml`. It builds the server, pushes `ghcr.io/tomknightdev/islanders-server:master` and redeploys the container on a DigitalOcean droplet over SSH. In practice, merging to master deploys to production. The Dockerfile copies only `server/` and `resources/`, so the server must never import anything from `client/`.

## Architecture

**Protocol.** Every message is a JSON `resources.Message{MessageType, Contents, ClientId}`. `Contents` is an `interface{}`, so the receiving side gets `map[string]interface{}` / `[]interface{}` with float64 numbers and unpacks it by hand with type assertions (for example `messageContents["pos"].([]interface{})`, or `resources.WorldMapWebSocketMessageConvert`). If you add a field to a `*Contents` struct, update the manual decoding on the receiving side too.

**Connection flow.** The client dials `ws://<addr>/connect` and sends `ConnectRequestMessage`. The server either creates or re-authenticates the account (stored in memory only, plaintext password, keyed by username). It replies with `ConnectResponseMessage`, which carries the client ID, the spawn position and the **entire generated world map**. It then sends `UpdateMessage`s for the other players and an NPC snapshot. Only after the connect response arrives does the client build the camera, `Player`, `World` and the chat and info windows (`handleConnectResponse` in `client/client.go`).

**Authority.** The client decides its own movement. `Player.Update` checks tile collision locally and sends its position, and the server stores it and relays it as-is. The server is authoritative only for NPCs. `npcLoop` ticks every 25ms and broadcasts `ServerEntityUpdateMessage`.

**Entity IDs share one uint16 space on the client.** Players get `0..n` (assigned as `len(clientsById)`), the unused enemies get 100+ and NPCs get 200+. The client puts players and NPCs into the same `client.NetworkPlayers` / `g.Entities` maps as `NetworkPlayer`s.

**World.** `server/worldgen.go` builds a 100×100 grid from rings (deep water, shallow water, sand, grass) plus random trees, and keeps a clear area around the spawn tile (50,50), which is pixel (400,400). Tiles are 8px. The map is stored as a single row-major layer, `Data[y*width+x]`. Tile constants and `IsPassable` live in `resources/resources.go`. The client's `World.Draw` converts a tile index into coordinates in `environments.png` as `((idx%16)-1)*8, (idx/16)*8`.

**Collision** checks the four corners of an 8×8 box against `IsPassable`. The same check exists twice, in `client/entities/player.go` (`canMoveTo`) and `server/npc.go` (`npcCanMoveTo`). Change both if the rules change.

**Rendering.** Every drawable entity shares one `ebiten-camera` instance. Each entity draws onto `cam.Surface` and calls `cam.Blit(screen)`, and `World.Draw` clears the surface first. Draw order is Environment → Entities → Player → Gui (`client/main.go`). All the GUI windows share one imgui `renderer.Manager`, and the windows are opened between `BeginFrame` and `EndFrame` in `Game.Update`.

**Client concurrency.** The GUI talks to the network over channels. The main menu sends server, username and password on `mm.Connect`, chat sends on `chat.SendChan`, and the player sends on `Player.SendChan`. Goroutines forward all of these into `client.SendChan`, and `connectToServer` writes them to the socket. These channels are unbuffered, so a send from `Update` blocks the game loop.

## Known pitfalls

- **Locking.** The server's client maps have no lock, yet connection goroutines and `npcLoop` both use them. Writes to a socket are protected by a per-client `c.mu`, but `sendUpdateToClients` returns without unlocking when a write fails. On the client, the network goroutine changes `g.Entities` while Ebiten iterates over it.
- **Dead code.** `serverLoop` and `server/entity.go` (the enemy chasers) are never called.
- **Spawn position** is hardcoded in several places in `server/server.go`. Keep them in sync with the clear area in `worldgen.go`.
