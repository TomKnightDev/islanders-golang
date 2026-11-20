# Islanders design

## Core loop

Islanders is a semi-idle multiplayer gathering and crafting game. The player picks a job by clicking on something in the world. Their character then works that job on its own until it can't continue, or until the player gives it something else to do.

```
mine iron ore ──► backpack full ──► walk to bank ──► deposit ──► walk back ──► mine…
                                                         │
bank ore ──► blacksmith smelts ore into ingots ──► ingots to bank
                                                         │
bank ingots ──► craft equipment at the anvil ──► better equipment makes jobs faster
```

### Mining
- Clicking an iron rock walks the character to a free tile next to it and starts mining.
- Each swing takes a set time and adds one **iron ore** to the backpack.
- When the backpack is full, the character walks to the bank, deposits everything, walks back to the same rock and carries on.
- **One miner per rock.** A rock is claimed while someone mines it. Clicking a claimed rock sends the character to the nearest free rock instead. If every rock is taken or depleted, the character waits by the clicked rock until one frees up.
- **Rocks run out.** Each rock holds a random amount of ore. When it's empty it turns into a depleted rock and respawns after a delay. The miner moves on to the nearest free rock and keeps going.
- The loop stops when:
  - the player's bank account has no room for the next deposit, or
  - the player clicks something else (see Action queue).

### Bank
Every player has their own bank account. The bank building is shared, but deposits and withdrawals only touch the player's own account, and capacity is per account.

### Smelting
- Clicking the blacksmith starts a smelting loop:
  1. Withdraw iron ore from the bank, as much as fits in the backpack.
  2. Walk to the blacksmith.
  3. Smelt ore into **iron ingots**, one at a time.
  4. Walk back and deposit the ingots.
- The loop repeats until the bank has no ore left.

### Crafting
- The blacksmith's anvil opens a recipe list. Each recipe uses ingots from the bank and takes time at the anvil.
- Equipment goes into slots on the character and changes job stats.
- First recipes:

| Item | Cost | Effect |
|---|---|---|
| Iron pickaxe | 3 ingots | Mining is 25% faster |
| Iron backpack straps | 5 ingots | Backpack holds +10 |
| Iron helmet | 4 ingots | Cosmetic for now |

### Action queue
- A character has one **current action** (a looping job) and at most one **queued action**.
- Clicking a target while busy queues it. The current action stops at its next safe point (the end of a swing, a deposit, or arriving on a tile), then the queued action starts.
- Clicking open ground is a plain walk. It also counts as a queued action, so the player can always break out of a loop.

### Idle behaviour
The server owns the character and simulates it, so loops keep running while the player is disconnected. Logging back in shows whatever was earned in the meantime. This works because characters already stay in server memory after disconnecting. Once persistence lands, progress survives server restarts too.

## Starting numbers

These are tuning values, kept in one place in the server code.

| Value | Start |
|---|---|
| Walk speed | 4 tiles/s |
| Mining swing | 2 s per ore |
| Ore per rock | 5–10, random |
| Rock respawn | 30 s |
| Backpack size | 20 items |
| Bank account size | 200 items |
| Smelting | 1 ore → 1 ingot, 3 s |
| Server tick | 25 ms (same as the NPC loop today) |

## World

- **Layers.** The map has two layers: ground (layer 0) and objects (layer 1). Value `0` in the objects layer means empty. Trees and buildings live on the object layer, so the grass underneath still renders.
- **Objects.** Interactive things are listed in `WorldMap.Objects` with an id, a kind and a tile position. Each kind's sprite is a constant in `resources/resources.go`.
- **Town.** The bank, blacksmith and anvil sit just north of the spawn clearing. The iron rocks are a cluster to the south-east, about 10 tiles away.
- **Passability.** Any non-zero object tile blocks movement. Ground tiles block according to `resources.IsPassable`.

| Object | Sprite (environments.png) |
|---|---|
| Iron rock | 90 grey boulder |
| Depleted iron rock | 107 small rock |
| Bank | 92 stone house front |
| Blacksmith | 93 wooden hut front |
| Anvil | 175 anvil |

The sprite sheet has no dedicated iron ore art. The grey boulder stands in until there's something better.

## Architecture decisions

- **The server is authoritative for players.** Today the client moves itself and the server relays positions. That can't work for an idle game: progress has to happen while the client is closed, and inventories can't be trusted to the client. Clients send intents (`walk to tile`, `interact with object`). The server pathfinds, moves the character each tick, runs the jobs and broadcasts positions and inventories. This is the same model NPCs already use.
- **One world goroutine owns all game state.** Connection handlers send it commands over a channel instead of touching shared maps. This also fixes the current data races.
- **Pathfinding** is A* over tiles, 4-directional, using the passability rule above.
- **Click-to-move replaces WASD.** The camera's `GetCursorCoords` turns a click into a world tile.
- **Typed messages.** New message contents are decoded with `resources.DecodeContents` (re-marshal into a struct) instead of hand-written `map[string]interface{}` assertions.

## Phases

Each phase is one PR and leaves the game playable.

1. **World objects.** Object layer, town and iron rocks in world generation, shared passability check, fix for last-column tiles, pre-rendered map. *(done)*
2. **Server-authoritative movement.** World goroutine and command channel, click-to-move with A*, client renders the server's position, WASD removed.
3. **Mining and inventory.** Action system with the queue rules above, the mining loop, rock claiming, depletion and respawn, backpack and per-player bank accounts, an inventory window, and new messages for inventory, action state and rock state.
4. **Smelting.** Blacksmith loop, ingots.
5. **Crafting and equipment.** Anvil recipe window, equipment slots and stat effects.
6. **Persistence.** Save characters, inventories and banks to disk on a Docker volume, and hash passwords. Until this lands, every deploy wipes progress.

## Decisions

- Bank accounts are private to each player.
- Iron rocks run out and respawn.
- Only one player can mine a rock at a time.
