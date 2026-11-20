# Islanders design

## Core loop

Islanders is a semi-idle multiplayer gathering and crafting game. The player explores the island to find villages and mines, then plans up to five actions in the **action window**. The character works through the list on its own, even while the player is offline.

```
mine iron ore ──► backpack full ──► nearest bank ──► deposit ──► back to the rock ──► mine…
                                                         │
bank ore ──► carry to a blacksmith ──► smelt into bars ──► bars to the bank
                                                         │
bank bars ──► craft equipment at the anvil ──► better equipment makes jobs faster
```

### The action window
- **Finding places.** A village or mine is *found* once the player walks into it (within its radius plus 2 tiles). Everyone starts having found their spawn village. Found places are listed with the actions available there:
  - a mine offers **Mine iron**
  - a village offers **Smelt iron** at its blacksmith
- **The action list** holds up to 5 actions. Clicking an action adds it to the end. Entries can be moved up or down or removed.
- **Goals.** Each action has one goal, which the player can change at any time:
  - **Until the bank has N** of the action's product (ore for mining, bars for smelting). New actions start at 100 ore or 50 bars.
  - **For N minutes** of the character's time. The timer pauses while the list is stopped, and runs at 10% speed while the player is offline like everything else.
- **Running the list.** **Start** works through the list from the top. When an action's goal is met, or its time is up, the character first banks anything it's carrying, then moves on. When the list is empty the character goes idle.
- **Stop** halts the list where it is; anything carried stays in the backpack. Reordering or removing the running action stops it and starts whatever is now first.
- **Clicks never cancel the list.** While the list is running, clicking the map does nothing. While it's stopped, clicking walks, which is how the player explores and finds new places.
- **Failures.** If the bank account is full, or there's no way to a place, the list stops and the action stays in it, so the player can sort things out and press Start again.

### Mining
- Mines only at the action's mine. The character walks to a free rock there and mines one **iron ore** per swing.
- When the backpack is full, or when it holds enough to reach the goal, the character walks to the nearest bank, deposits, and goes back to the same rock.
- **One miner per rock.** A rock is reserved when a miner is sent to it. If every rock in the mine is taken or depleted, the character waits at the mine until one frees up.
- **Rocks run out.** Each rock holds a random amount of ore. When it's empty it turns into a depleted rock and respawns after a delay. The miner moves on to the nearest free rock in the same mine.

### Smelting
- Smelts at the action's village blacksmith, in trips:
  1. At the nearest bank: deposit anything carried, then take out iron ore, as much as fits in the backpack but no more than the goal still needs.
  2. Walk to the blacksmith and smelt one **iron bar** per ore, one at a time.
  3. Walk back to the bank, deposit the bars, and start the next trip.
- If the bank runs out of ore, the action ends early ("No ore to smelt").

### Items and stacking
Backpacks and bank accounts are made of slots.

- **Backpack:** items don't stack. Every item takes its own slot, so a player carries at most 20 items (20 iron ore per trip).
- **Bank:** each slot holds one stack of a single item kind, up to that item's stack limit.

| Item | Bank stack limit |
|---|---|
| Iron ore | 10 |
| Iron bar | 20 |

A container is full for an item when every slot is used and every stack of that item is at its limit.

### Bank
Every village has a bank building, and every player has their own bank account, shared across all banks. Deposits and withdrawals only touch the player's own account, and capacity is per account.

### Crafting
- The blacksmith's anvil opens a recipe list. Each recipe uses ingots from the bank and takes time at the anvil.
- Equipment goes into slots on the character and changes job stats.
- First recipes:

| Item | Cost | Effect |
|---|---|---|
| Iron pickaxe | 3 ingots | Mining is 25% faster |
| Iron backpack straps | 5 ingots | Backpack holds +10 |
| Iron helmet | 4 ingots | Cosmetic for now |

### Activity and progress bar
The character shows what it's doing ("Mining", "Smelting", "Walking to the bank", "Waiting for a free rock", "Bank full") in a bar at the bottom middle of the screen. The bar fills green over each timed step: every mining swing, every smelt, every walk. Open-ended states show an empty bar, and the bar hides when idle.

### Idle behaviour
The server owns the character and simulates it, so loops keep running while the player is disconnected, at 10% speed. Walking and every timed step run slower together, because each character has its own clock that only advances every tenth tick while they're offline. Logging back in shows whatever was earned in the meantime. Offline characters stay visible to other players while they work. This works because characters already stay in server memory after disconnecting. Once persistence lands, progress survives server restarts too.

## Starting numbers

These are tuning values, kept in one place in the server code.

| Value | Start |
|---|---|
| Walk speed | 4 tiles/s |
| Mining swing | 2 s per ore |
| Ore per rock | 5–10, random |
| Rock respawn | 30 s |
| Backpack size | 20 slots |
| Bank account size | 100 slots |
| Smelting | 1 ore → 1 bar, 3 s |
| Offline speed | 10% |
| Server tick | 25 ms |

## World

The island is generated by `server/worldgen.go` from a fixed seed, so it's the same on every start and deploy and a shared location always points to the same spot. It's 160×160 tiles.

- **Island.** Noise shaped by a falloff from the centre gives an irregular coastline: deep sea, shallow water, a sandy beach, then grassland. The map edge is always open sea.
- **Rivers** start in the highlands and wander downhill to the sea. They're one tile wide up high and two wide lower down, and where they meander over flat ground they spread into small lakes. Rivers can't be walked through.
- **Villages** (4): a cobbled square with the bank, blacksmith and anvil on its north side and cottages around it. The village nearest the middle, Oakhaven, has a tower and is where everyone spawns.
- **Mines** (3): a patch of stony ground with six iron rocks, on higher ground away from villages.
- **Roads** link the villages (a minimum spanning tree) and run from each mine to its nearest village. They prefer open ground, merge where they meet, clear trees in their way, and cross rivers on wooden bridges.
- **Forests** grow where a second noise field is high: dense woods on a darker forest floor with mushrooms, thinning out at the edges, with some autumn-coloured patches. Elsewhere there are scattered trees, bushes, flower meadows and pebbles.
- **Reachability.** After generation, every bank, blacksmith, anvil and rock is checked to be walkable from the spawn; anything cut off would get a road.

### Layers
| Layer | Holds | Blocks movement |
|---|---|---|
| 0 Ground | grass, sand, water, roads, bridges, cobbles, forest floor | water only |
| 1 Objects | trees, bushes, buildings, rocks | always |
| 2 Decor | flowers, pebbles, mushrooms, stones, rubble | never |

### Places
Villages and mines are named and listed in `WorldMap.Places`. The info panel describes a location as "in Oakhaven", "at Ironridge Mine" or "near Saltmere", and the names float over the map.

### Sprites
| Thing | Sprite (environments.png) |
|---|---|
| Iron rock / depleted | 90 grey boulder / 107 small rock |
| Bank | 92 stone house front |
| Blacksmith | 93 wooden hut front |
| Anvil | 175 anvil |
| Cottage / tower | 95 / 94 |
| River | 115 water |
| Road / bridge / square | 191 dirt / 13 planks / 108 cobblestone |
| Trees | 74, 78, 80, 75 (autumn), 79 (bush) |
| Decor | 111, 112 flowers; 119 pebbles; 109 mushroom; 118 stones; 123 rubble |

The sprite sheet has no dedicated iron ore art. The grey boulder stands in until there's something better.

## Architecture decisions

- **The server is authoritative for players.** Today the client moves itself and the server relays positions. That can't work for an idle game: progress has to happen while the client is closed, and inventories can't be trusted to the client. Clients send intents (`walk to tile`, `interact with object`). The server pathfinds, moves the character each tick, runs the jobs and broadcasts positions and inventories.
- **One world goroutine owns all game state.** Connection handlers send it commands over a channel instead of touching shared maps. This also fixes the current data races.
- **Pathfinding** is A* over tiles, 8-directional, using the passability rule above. A diagonal step needs both tiles beside it free, so sprites never clip a corner.
- **Click-to-move replaces WASD.** The camera's `GetCursorCoords` turns a click into a world tile.
- **Typed messages.** New message contents are decoded with `resources.DecodeContents` (re-marshal into a struct) instead of hand-written `map[string]interface{}` assertions.

## Locations

Players share where they are with a grid reference instead of raw coordinates. The island is split into 10×10-tile squares, lettered A–P west to east and numbered 1–16 north to south. A location reads "H8 (78, 79)": the square to say out loud, then the exact tile, followed by the place: "in Oakhaven", "at Ironridge Mine" or "near Saltmere". The info panel shows the player's location and has a **Share location** button that posts it to chat.

## NPCs

The wandering NPCs and their random chat were removed in phase 3 to keep the world quiet. They can come back later with a purpose (a banker, a blacksmith).

## Phases

Each phase is one PR and leaves the game playable.

1. **World objects.** Object layer, town and iron rocks in world generation, shared passability check, fix for last-column tiles, pre-rendered map. *(done)*
2. **Server-authoritative movement.** World goroutine and command channel, click-to-move with A*, client renders the server's position, WASD removed. *(done)*
3. **Mining and inventory.** Action system with the rules above, the mining loop, rock claiming, depletion and respawn, backpack and per-player bank accounts, an inventory window, and new messages for inventory, action state and rock state. *(done)*
4. **Smelting and the action window.** Found places, a 5-action list with amount or time goals, smelting trips. *(done)*
5. **Crafting and equipment.** Anvil recipe window, equipment slots and stat effects.
6. **Persistence.** Save characters, inventories and banks to disk on a Docker volume, and hash passwords. Until this lands, every deploy wipes progress.

## Decisions

- Bank accounts are private to each player.
- Iron rocks run out and respawn.
- Only one player can mine a rock at a time.
