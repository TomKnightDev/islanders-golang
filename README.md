# Islanders

A small multiplayer, semi-idle, top-down 2D game written in Go. Explore an island, find its villages, mines, forests, farms and camps, then plan a list of actions for your character to work through on its own, even while you're offline.

## The game

- **Classes.** Create a warrior, archer, mage or necromancer. Each starts with its own gear and combat skills.
- **The action list.** Plan up to five actions: mine, chop, farm, smelt, fletch, fight, or visit a village building. Each has a goal of 10, 20, 50 or 100. Adding an action to an empty list starts it. Your character walks the roads, banks what it gathers and moves on to the next action.
- **Skills.** Twelve skills, levels 1–99: Melee, Strength, Defence, Ranged, Magic, Hitpoints, Mining, Woodcutting, Farming, Smithing, Fletching and Stamina. Fight XP is only paid when the enemy dies.
- **Stamina.** Work tires your character out; they rest at the nearest inn (or the free, slower stable if they can't pay) and then carry on.
- **Items and quality.** Gear, tools and materials have a quality from 1 to 100. Better skill and better tools gather better materials, and better gear hits harder and sells for more. A tool belt holds a pickaxe, hatchet, sickle, hammer and knife.
- **Villages.** Each village has a bank, a store, an inn, a blacksmith and a fletcher. Visit the bank to move items and equip gear (its contents are only visible there), the store to buy and sell, and the inn for rumours of places you haven't found.
- **Coins.** Earned from camp loot and selling to stores, spent at stores and inns. Store prices fall as they buy more of something and recover over time.

The island is generated from a fixed seed, so it's the same every time the server starts. There's no persistence yet: characters live in server memory and are lost when it restarts.

More detail is in [`docs/DESIGN.md`](docs/DESIGN.md) (game design and phase plan) and [`docs/ECONOMY.md`](docs/ECONOMY.md) (the economy plan).

## Requirements

- Go (see `go.mod` for the version).
- The client uses cgo, so it needs a C compiler (GCC or Clang). Building the Windows client from macOS needs mingw-w64 (`brew install mingw-w64`).

## Running locally

```bash
make run-server   # dev server on 0.0.0.0:8285 with dev tools enabled
make run-client   # the desktop client
```

In the client, choose **Play locally** to connect to `localhost:8285`, or **Play online** for the production server. Pick a name and password; a new name takes you to class selection.

With a dev server, press **F12** in the client to open the Dev window and speed the game up (1×–20×) for testing.

A client and server must come from the same version: protocol changes mean an old client can't play a new server.

## Building

```bash
make build                  # bin/server and bin/client
make build-client-windows   # bin/windows/Islanders.exe
```

To run the server in Docker:

```bash
docker build -t islanders-server .
docker run -d -p 8285:8285 islanders-server
```

## Tests

```bash
go vet ./...
go test ./server/ ./resources/
```

## Project layout

- `server/`: the websocket game server. It owns world generation, every character's movement, jobs and fights, the enemies, stores and banks.
- `client/`: the desktop client, built on Ebiten with imgui windows.
- `resources/`: message types, items, skills and the world map, shared by both.

## Deployment

Every push to `master` builds the server image, pushes it to `ghcr.io/tomknightdev/islanders-server:master` and redeploys it, so merging to `master` deploys to production. CI also builds the Windows client and uploads it as the `Islanders-windows` artifact.
