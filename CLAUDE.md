# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Islanders is a small multiplayer top-down 2D game. One Go module (`github.com/tomknightdev/islanders-golang`) holds three packages:

- `server/`: a websocket game server (gorilla/websocket). It owns world generation, the player registry and every character's movement and jobs. There are no NPCs at the moment; the old wandering NPCs were removed and are in git history (`server/npc.go`).
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

Without `-addr`, the server binds to `<outbound IP>:8285`. The client's main menu has two buttons: **Play online** connects to the production server and **Play locally** to `localhost:8285` (both addresses are constants in `client/gui/mainmenu.go`). `make run-server` listens on all interfaces, so Play locally reaches it.

## Deployment

Every push to `master` triggers `.github/workflows/server.yml`. It builds the server, pushes `ghcr.io/tomknightdev/islanders-server:master` and redeploys the container on a DigitalOcean droplet over SSH. In practice, merging to master deploys to production. The Dockerfile copies only `server/` and `resources/`, so the server must never import anything from `client/`.

## Architecture

**Protocol.** Every message is a JSON `resources.Message{MessageType, Contents, ClientId}`. `Contents` is an `interface{}`, so the receiving side gets `map[string]interface{}` / `[]interface{}` with float64 numbers. Older handlers unpack it by hand with type assertions (for example `messageContents["pos"].([]interface{})`); if you add a field to one of those `*Contents` structs, update that manual decoding too. New code should use `resources.DecodeContents(contents, &typedStruct)` instead, as the world map already does.

**Connection flow.** The client dials `ws://<addr>/connect` and sends `ConnectRequestMessage`. The server either creates or re-authenticates the account (stored in memory only, plaintext password, keyed by username). Logging in from a second connection closes the first. The server replies with `ConnectResponseMessage`, which carries the client ID, the player's position and the **entire generated world map**. It then sends `UpdateMessage`s for the other players and the player's inventory. Only after the connect response arrives does the client build the camera, `Player`, `World`, the action bar and the chat, info and inventory windows (`handleConnectResponse` in `client/client.go`).

**Authority.** The server owns every character. A click sends `MoveRequestMessage` with a target tile. The server pathfinds (A*, 8-directional with no corner-cutting past blocked tiles, `server/pathfind.go`); clicking a blocked tile walks to a free tile next to it. The server then moves the character `walkStep` pixels per 25ms tick and broadcasts `UpdateMessage` to **every** client, the owner included. The client never moves itself; `receiveUpdateMessage` applies the owner's position to `client.Player`. Characters stay in the world while their player is offline and keep walking, and other clients still see them.

**Jobs and the action list** (`server/jobs.go`). Each player has an action list (`player.actions`, up to `resources.MaxActions`) edited by `ActionCommandMessage`s (add, edit, remove, up, down, start, stop) through `World.actionCommand`, which validates that the place has been found (`player.found`, filled by `tickDiscovery`) and fits the action. While `player.running`, `tickJob` starts the first action as a `job`: a small state machine (mining: walk to rock → mine → bank → back; smelting: bank (deposit, collect ore) → blacksmith → smelt → bank). Goals are checked at decision points (`amountMet`) and every tick for time goals (`queuedAction.ran` counts only running ticks); `wrapUp` banks what's carried before `finishAction` moves on, and `failAction` stops the list but keeps the action. Clicks (`World.click`) only walk, and are ignored while the list runs. Rocks belong to a mine (`rock.place`), are claimed when a miner is sent to them (one miner per rock), deplete to `TileIronRockDepleted` and respawn on a timer; tile changes go to clients as `ObjectUpdateMessage`. Backpacks and per-player bank accounts are `resources.Inventory` (slots of stacks). The bank stacks up to `resources.StackLimits`; the backpack is created with `Stacking: false`, so every item takes its own slot. Changes set `player.invDirty`, and the tick sends the owner an `InventoryMessage` with backpack, bank, activity, found places and the action list. All tuning values are constants at the top of `jobs.go`. Each player has their own clock (`player.ticks`), which advances every tick online and every `offlineSlowdown` ticks offline; job and action timing use it, while world timers such as rock respawns use `World.ticks`.

**Action bar.** Every activity change or new timed step (each mining swing, smelt or walk) calls `startAction`, which sends the owner the action's length and elapsed time in the `InventoryMessage`. The client's `gui.ActionBar` animates the green fill locally from that, so the server doesn't send progress every tick. Open-ended activities (idle, waiting, bank full) have length 0 and show an empty bar; idle hides it.

**Sending state to writers.** A message handed to `session.push` is encoded later on the writer goroutine, so it must not share memory the world keeps changing. Use `Inventory.Clone()` and `WorldMap.Clone()` when putting live state into a message.

**Entity IDs.** Players get `0, 1, 2…` from `World.nextID`. Other players are `NetworkPlayer`s in `client.NetworkPlayers` / `g.Entities`. `ServerEntityUpdateMessage` and its client handler are left over from NPCs and currently unused.

**World.** `server/worldgen.go` generates a 160×160 island from a fixed seed (`worldSeed`), so the map is identical on every start; changing the generator or seed moves everything players know. Pipeline: island elevation from value noise with a radial falloff → rivers walked downhill from highlands → terrain painted from elevation bands → villages and mines placed on the mainland away from water → roads (weighted A*, bridges over rivers) → forests and decoration → a reachability check from the spawn. Three row-major layers (`Data[y*width+x]`): ground (opaque; water blocks), objects (anything non-zero blocks) and decor (never blocks). Interactive objects are listed in `WorldMap.Objects`, named villages and mines in `WorldMap.Places`, and the spawn in `WorldMap.SpawnX/Y`. Tiles are 8px; values are Tiled-style gids into `environments.png` (atlas column `(gid-1)%16`, row `(gid-1)/16`). The client's `World` pre-renders all layers into one image and draws place names over the map. To look at the island without running the game, render `generateWorld(worldSeed)` to a PNG from a throwaway test.

**Collision** is `WorldMap.TilePassable` / `BoxPassable` in `resources`, used by server pathfinding. A ground tile blocks if `IsPassable` says so; any non-empty object tile blocks.

**Rendering.** Every drawable entity shares one `ebiten-camera` instance. Each entity draws onto `cam.Surface` and calls `cam.Blit(screen)`, and `World.Draw` clears the surface first. Draw order is Environment → Entities → Player → Gui (`client/main.go`). All the GUI windows share one imgui `renderer.Manager`, and the windows are opened between `BeginFrame` and `EndFrame` in `Game.Update`.

**Server concurrency.** `World` (`server/world.go`) owns all game state, and only its `Run` goroutine touches it. Connection handlers (`server/server.go`) only parse messages and submit closures with `world.Do(func(w *World){...})`; anything needing a reply passes a channel inside the closure (see login). Each connection is a `session` whose writer goroutine is the only thing writing to the socket. The world queues messages with `session.push`, which never blocks: a client whose 256-message buffer fills is disconnected.

**Client concurrency.** The network reader goroutine only queues messages on `client.RecvChan`. `Game.Update` drains the queue (`handleMessages`), so game state is only touched from the Ebiten loop. Outbound, the main menu sends server, username and password on `mm.Connect`, chat sends on `chat.SendChan`, and the player sends move requests on `Player.SendChan`, which is buffered and sent without blocking. Goroutines forward all of these into `client.SendChan`, and `connectToServer` writes them to the socket.

## Known pitfalls

- **No persistence.** Accounts, positions and everything else live in server memory, and every deploy wipes them.
- **Older clients** still send `UpdateMessage` with their own position; the server ignores it.
