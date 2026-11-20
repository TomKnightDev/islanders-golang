# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Islanders is a small multiplayer top-down 2D game. One Go module (`github.com/tomknightdev/islanders-golang`) holds three packages:

- `server/`: a websocket game server (gorilla/websocket). It owns world generation, the player registry and the NPCs.
- `client/`: a desktop client built on Ebiten v2, with imgui for menus and chat. It needs **cgo/GCC** to build.
- `resources/`: shared message types, tile constants and the `WorldMap` struct. Both sides import it.

Game design, the planned gathering/crafting loop and the phase plan live in `docs/DESIGN.md`. Read it before working on gameplay.

## Commands

```bash
make build          # builds bin/server and bin/client
make run-server     # go run ./server/ -addr 0.0.0.0:8285 (override with SERVER_ADDR=...)
make run-client     # go run ./client/
go build ./server/  # what CI runs; the client is not built in CI
go vet ./...
go test ./server/ ./resources/   # what CI runs; client/client_test.go is commented out
go test ./client -run TestName   # single test
```

Without `-addr`, the server binds to `<outbound IP>:8285`. The client's main menu defaults to the production server (`134.209.26.73:8285`). Change the address to `localhost:8285` to test locally.

## Deployment

Every push to `master` triggers `.github/workflows/server.yml`. It builds the server, pushes `ghcr.io/tomknightdev/islanders-server:master` and redeploys the container on a DigitalOcean droplet over SSH. In practice, merging to master deploys to production. The Dockerfile copies only `server/` and `resources/`, so the server must never import anything from `client/`.

## Architecture

**Protocol.** Every message is a JSON `resources.Message{MessageType, Contents, ClientId}`. `Contents` is an `interface{}`, so the receiving side gets `map[string]interface{}` / `[]interface{}` with float64 numbers. Older handlers unpack it by hand with type assertions (for example `messageContents["pos"].([]interface{})`); if you add a field to one of those `*Contents` structs, update that manual decoding too. New code should use `resources.DecodeContents(contents, &typedStruct)` instead, as the world map already does.

**Connection flow.** The client dials `ws://<addr>/connect` and sends `ConnectRequestMessage`. The server either creates or re-authenticates the account (stored in memory only, plaintext password, keyed by username). Logging in from a second connection closes the first. The server replies with `ConnectResponseMessage`, which carries the client ID, the player's position and the **entire generated world map**. It then sends `UpdateMessage`s for the other players and an NPC snapshot. Only after the connect response arrives does the client build the camera, `Player`, `World` and the chat and info windows (`handleConnectResponse` in `client/client.go`).

**Authority.** The server owns every character. A click sends `MoveRequestMessage` with a target tile. The server pathfinds (A*, 4-directional, `server/pathfind.go`); clicking a blocked tile walks to a free tile next to it. The server then moves the character `walkStep` pixels per 25ms tick and broadcasts `UpdateMessage` to **every** client, the owner included. The client never moves itself; `receiveUpdateMessage` applies the owner's position to `client.Player`. Characters stay in the world while their player is offline and keep walking, and other clients still see them.

**Entity IDs share one uint16 space on the client.** Players get `0, 1, 2…` from `World.nextID` and NPCs get 1000+ (`npcIDBase`). The client puts players and NPCs into the same `client.NetworkPlayers` / `g.Entities` maps as `NetworkPlayer`s.

**World.** `server/worldgen.go` builds a 100×100 map with two row-major layers (`Data[y*width+x]`): ground (rings of deep water, shallow water, sand, grass) and objects (trees plus a fixed town: bank, blacksmith, anvil, iron rocks). `0` in the object layer means empty. Interactive objects are also listed in `WorldMap.Objects` with an id, kind and tile position. A clear area surrounds the spawn tile (50,50), which is pixel (400,400), and a ring around every object is kept free of trees. Tiles are 8px. Tile and object constants live in `resources/resources.go`. Tile values are Tiled-style gids into `environments.png`: atlas column `(gid-1)%16`, row `(gid-1)/16`. The client's `World` pre-renders all layers into one image the first time it draws.

**Collision** is `WorldMap.TilePassable` / `BoxPassable` in `resources`, shared by the client player and server NPCs. A ground tile blocks if `IsPassable` says so; any non-empty object tile blocks.

**Rendering.** Every drawable entity shares one `ebiten-camera` instance. Each entity draws onto `cam.Surface` and calls `cam.Blit(screen)`, and `World.Draw` clears the surface first. Draw order is Environment → Entities → Player → Gui (`client/main.go`). All the GUI windows share one imgui `renderer.Manager`, and the windows are opened between `BeginFrame` and `EndFrame` in `Game.Update`.

**Server concurrency.** `World` (`server/world.go`) owns all game state, and only its `Run` goroutine touches it. Connection handlers (`server/server.go`) only parse messages and submit closures with `world.Do(func(w *World){...})`; anything needing a reply passes a channel inside the closure (see login). Each connection is a `session` whose writer goroutine is the only thing writing to the socket. The world queues messages with `session.push`, which never blocks: a client whose 256-message buffer fills is disconnected.

**Client concurrency.** The network reader goroutine only queues messages on `client.RecvChan`. `Game.Update` drains the queue (`handleMessages`), so game state is only touched from the Ebiten loop. Outbound, the main menu sends server, username and password on `mm.Connect`, chat sends on `chat.SendChan`, and the player sends move requests on `Player.SendChan`, which is buffered and sent without blocking. Goroutines forward all of these into `client.SendChan`, and `connectToServer` writes them to the socket.

## Known pitfalls

- **No persistence.** Accounts, positions and everything else live in server memory, and every deploy wipes them.
- **Older clients** still send `UpdateMessage` with their own position; the server ignores it.
