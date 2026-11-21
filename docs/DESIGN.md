# Islanders design

## Core loop

Islanders is a semi-idle multiplayer gathering and crafting game. The player explores the island to find villages and mines, then plans up to five actions in the **action window**. The character works through the list on its own, even while the player is offline.

```
gather (mine ore / chop logs) ──► backpack full ──► nearest bank ──► back to the rock or tree ──► …
                                                         │
bank ore  ──► blacksmith ──► iron bars ──► bank
bank logs ──► fletcher's hut ──► arrows ──► bank
                                                         │
bank bars ──► craft equipment at the anvil ──► better equipment makes jobs faster
```

### The action window
- **Finding places.** A village, mine, named forest or farm is *found* once the player walks into it (within its radius plus 2 tiles). Everyone starts having found their spawn village.
- **Adding actions.** The Actions window holds just the list and its buttons: **Add action**, **Start**/**Stop**. **Add action** opens a separate picker window, which stays open while you add several. It shows one row per action (Mine, Chop, Farm, Smelt, Fletch), at the **nearest found place** for it, with its distance and a button per material. Materials the player doesn't have the level for are greyed out with the level needed. Actions with no place found yet say what to look for.
- **The action list** holds up to 5 actions. Clicking an action adds it to the end. Entries can be moved up or down or removed.
- **Goals.** Each action has one goal, which the player can change at any time:
  - **Until the bank has N** of the action's product. New actions start at 100 (ore, logs, crops), 50 bars or 500 arrows.
  - **For N minutes** of the character's time. The timer pauses while the list is stopped, and runs at 10% speed while the player is offline like everything else.
- **Running the list.** **Start** works through the list from the top. When an action's goal is met, or its time is up, the character first banks anything it's carrying, then moves on. When the list is empty the character goes idle.
- **Stop** halts the list where it is; anything carried stays in the backpack. Reordering or removing the running action stops it and starts whatever is now first.
- **Clicks never cancel the list.** While the list is running, clicking the map does nothing. While it's stopped, clicking walks, which is how the player explores and finds new places.
- **Failures.** If the bank account is full, or there's no way to a place, the list stops and the action stays in it, so the player can sort things out and press Start again.

### Skills
The game is meant to be played for hundreds of hours, so skills level slowly, with lots to unlock on the way.

- **Gathering and crafting skills:** Mining, Woodcutting, Farming, Smithing (smelting), Fletching. Each action trains one. The five combat skills are under Combat.
- **Levels 1–99.** The XP curve roughly doubles every seven levels: level 2 at 83 XP, level 50 at about 100,000, level 99 at about 13 million.
- **XP** comes from every item gathered, and every input crafted. Higher tiers give more.
- **Speed.** Each level takes 0.4% off that skill's step time, so level 99 works in about 61% of the base time.
- **Tiers.** Levels unlock materials:

| Skill | Level 1 | 15 | 30 | 45 |
|---|---|---|---|---|
| Mining | copper (10 XP) | iron (25) | gold (45) | mithril (80) |
| Woodcutting | oak (10) | pine (25) | maple (45) | dark oak (80) |
| Smithing | copper bars (12) | iron bars (28) | gold bars (50) | mithril bars (90) |
| Fletching | oak arrows (12) | pine arrows (28) | maple arrows (50) | dark oak arrows (90) |
| Farming | cabbages (8) | cotton at level 10 (18) | | |

- **Level-ups** post a congratulations message in the player's chat.
- The **Skills window** shows each level with a progress bar to the next, with the XP figures centred on the bar.

### Gathering: mining, chopping and farming
All three work the same way, from *nodes* of the chosen material at the action's place.

| | Mining | Chopping | Farming |
|---|---|---|---|
| Where | a mine | a named forest | a farm |
| Base speed | 1 ore per 4 s | 1 log per 6 s | 1 crop per 5 s |
| A node holds | 5–10 ore | 1–4 logs | 2–5 crops |
| When empty | depleted rock, 30 s | felled tree, 60 s | harvested plot, 45 s |

Mines are ranked by distance from the spawn: the nearest has copper and iron, the next iron and gold, the farthest gold and mithril. Forests have whichever trees grew there. Every farm has rows of cabbages and cotton. Only trees inside a named forest can be chopped; scattered trees are scenery. A node with no walkable tile beside it is skipped.

### Mining
- Mines only at the action's mine. The character walks to a free rock there and mines one **iron ore** per swing.
- When the backpack is full, or when it holds enough to reach the goal, the character walks to the nearest bank, deposits, and goes back to the same rock.
- **One miner per rock.** A rock is reserved when a miner is sent to it. If every rock in the mine is taken or depleted, the character waits at the mine until one frees up.
- **Rocks run out.** Each rock holds a random amount of ore. When it's empty it turns into a depleted rock and respawns after a delay. The miner moves on to the nearest free rock in the same mine.

### Crafting at a station: smelting and fletching
Both run in trips from the action's village:
1. At the nearest bank: deposit anything carried, then take out the input, as much as fits in the backpack but no more than the goal still needs.
2. Walk to the station and work through the input one at a time.
3. Walk back to the bank, deposit the output, and start the next trip.

| | Smelting | Fletching |
|---|---|---|
| Station | blacksmith | fletcher's hut |
| Makes | 1 bar from 1 ore, 6 s | 10 arrows from 1 log, 4 s |
| Ends early | "No ore to smelt" | "No logs to fletch" |

### Combat
- **Five combat skills**: Attack (melee accuracy), Strength (melee max hit), Defence (how often enemies hit you), Ranged (bow accuracy and max hit) and Hitpoints (max health). Hitpoints starts at level 10; max health is 10 × the Hitpoints level, so 100 to start.
- **XP.** Each point of melee damage gives 2 Attack and 2 Strength XP; ranged damage gives 4 Ranged XP; all damage gives 1⅓ Hitpoints XP; each enemy attack you block gives 2 Defence XP.
- **Formulas** are RuneScape-style, scaled 10× to match the health numbers: an attack roll `(level + 8) × (accuracy + 64)` against a defence roll `(level + 9) × (armour + 64)` sets the hit chance, and `(level + 8) × (strength + 64) / 64` is the max hit.
- **Health** regenerates slowly out of combat and quickly while resting at a bank. **Food** heals when eaten: a cabbage restores 30.
- **Defeat.** At 0 health you lose everything in the backpack (not your equipment), wake up in your spawn village at full health, and the action list stops.

### Equipment
- **Slots:** weapon, ammo, head, body, shield. The Equipment window shows them, your combat stats, and the gear in your bank; click to equip, Remove to put back. Gear above your level is greyed out with the level it needs.
- **Weapons:** a training sword (everyone starts with one); copper, iron, gold and mithril swords (Attack 1/15/30/45); oak, pine, maple and dark oak shortbows (Ranged 1/15/30/45). Bows are two-handed, attack from up to 5 tiles, and use one arrow from the ammo slot per shot; equipping arrows moves all of that kind from the bank. Better arrows hit harder.
- **Armour:** helmet, platebody and shield in each metal (Defence 1/15/30/45).
- **Starting kit:** a training sword and 5 cabbages in the bank.
- Gear comes from camp loot for now; crafting it at the anvil is a later phase.

### Camps and fighting
- **Four camps**, easiest nearest the spawn: Goblin Camp (goblins, level 3), Bandit Hideout (bandits, 12), Orc Warband (orc warriors, 28), Skeleton Ruins (skeletons, 45). Each is a dirt clearing with huts (ruins for the skeletons) and four enemies who wander near home, with health bars when hurt.
- **Fight** is an action at a found camp, with a goal of N kills or a time. Each trip: deposit at the bank (resting first below 60% health until 90%), take up to 10 food, walk to the camp, then fight enemy after enemy. One player fights an enemy at a time; it respawns 30 s after it dies.
- **In a fight** the character eats below half health, retreats to the bank to rest when below 30% with no food left, and banks the loot when the backpack is full. A bow user runs out of arrows and the list stops.
- **Loot** per kill rolls each drop in the enemy's table: their tier's ore and bars, food, arrows, and small chances of each piece of their tier's gear.
- Fights run on the fighter's clock, so an offline character's fights slow down with them.

### Items and stacking
Backpacks and bank accounts are made of slots.

- **Backpack:** most items don't stack. Every item takes its own slot, so a player carries at most 20 items (20 iron ore per trip). Arrows are the exception and stack 50 to a slot.
- **Bank:** each slot holds one stack of a single item kind, up to that item's stack limit.

| Item | Bank stack limit |
|---|---|
| Ores (each kind) | 10 |
| Bars (each kind) | 20 |
| Logs (each kind) | 20 |
| Arrows (each kind) | 50 (also 50 in the backpack) |
| Cabbage, cotton | 20 |

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

### Screen layout
Each time the game loads, the windows snap to their places: **Info** top left with **Skills** below it, **Actions** top right (the action picker opens beside it), **Inventory** bottom right, and **Chat** bottom left with its message box along the bottom (Enter sends). The progress bar sits bottom middle. Windows can be moved during a session.

### Camera and route
The camera follows the character. Click and drag the map to look around; a plain click still walks there. While the camera is free, a **Follow my character** button at the top of the screen brings it back. Your own route is drawn as a line from your character to where they're heading; nobody else sees it.

### Hover information
Pointing at a building, rock or tree on the map shows a tooltip with its name and what it's for, including where it is ("Saltmere's bank", "Iron rock at Old Quarry") and, for trees, whether they can be chopped. Items in the inventory window have tooltips too. The inventory window doesn't show the current activity; the progress bar does.

### Activity and progress bar
The character shows what it's doing ("Mining", "Smelting", "Walking to the bank", "Waiting for a free rock", "Bank full") in a bar at the bottom middle of the screen. The bar fills green over each timed step: every mining swing, every smelt, every walk. Open-ended states show an empty bar, and the bar hides when idle.

### Idle behaviour
The server owns the character and simulates it, so loops keep running while the player is disconnected, at 10% speed. Walking and every timed step run slower together, because each character has its own clock that only advances every tenth tick while they're offline. Logging back in shows whatever was earned in the meantime. Offline characters stay visible to other players while they work. This works because characters already stay in server memory after disconnecting. Once persistence lands, progress survives server restarts too.

## Starting numbers

These are tuning values, kept in one place in the server code.

| Value | Start |
|---|---|
| Walk speed | 2 tiles/s |
| Mining swing | 4 s per ore |
| Ore per rock | 5–10, random |
| Rock respawn | 30 s |
| Backpack size | 20 slots |
| Bank account size | 100 slots |
| Smelting | 1 ore → 1 bar, 6 s |
| Chopping | 1 log per 6 s; trees hold 1–4, regrow in 60 s |
| Fletching | 1 log → 10 arrows, 4 s |
| Farming | 1 crop per 5 s; plants hold 2–5, regrow in 45 s |
| Your attack / enemy attack | every 4.8 s / every 6 s |
| Enemy respawn | 30 s |
| Health regeneration | 1 per 1.2 s, 1 per 0.25 s resting at the bank |
| Offline speed | 10% |
| Server tick | 25 ms |

## World

The island is generated by `server/worldgen.go` from a fixed seed, so it's the same on every start and deploy and a shared location always points to the same spot. It's 160×160 tiles.

- **Island.** Noise shaped by a falloff from the centre gives an irregular coastline: deep sea, shallow water, a sandy beach, then grassland. The map edge is always open sea.
- **Rivers** start in the highlands and wander downhill to the sea. They're one tile wide up high and two wide lower down, and where they meander over flat ground they spread into small lakes. Rivers can't be walked through.
- **Villages** (4): a cobbled square with the bank, blacksmith and anvil on its north side, a fletcher's hut on its south side, and cottages around it. The village nearest the middle, Oakhaven, has a tower and is where everyone spawns.
- **Mines** (3): a patch of stony ground with six iron rocks, on higher ground away from villages.
- **Roads** link the villages (a minimum spanning tree) and run from each mine to its nearest village. They prefer open ground, merge where they meet, clear trees in their way, and cross rivers on wooden bridges.
- **Farms** (3): a tilled field a short walk from a village, with rows of cabbages and cotton, each row beside a walkable one. Roads lead to them and never cut through the crops.
- **Named forests** (4): the largest woods get a name (Whisperwood, Elderglade, Mossy Hollow, Thornwood) and become places where trees can be chopped.
- **Forests** grow where a second noise field is high: dense woods on a darker forest floor with mushrooms, thinning out at the edges, with some autumn-coloured patches. Elsewhere there are scattered trees, bushes, flower meadows and pebbles.
- **Reachability.** After generation, every bank, blacksmith, anvil and rock is checked to be walkable from the spawn; anything cut off would get a road.

### Layers
| Layer | Holds | Blocks movement |
|---|---|---|
| 0 Ground | grass, sand, water, roads, bridges, cobbles, forest floor | water only |
| 1 Objects | trees, bushes, buildings, rocks | always |
| 2 Decor | flowers, pebbles, mushrooms, stones, rubble | never |

### Places
Villages, mines, farms and the largest forests are named and listed in `WorldMap.Places`, with the materials each gathering place offers. The info panel describes a location as "in Oakhaven", "at Ironridge Mine" or "near Saltmere", and the names float over the map.

### Sprites
| Thing | Sprite (environments.png) |
|---|---|
| Copper / iron / gold / mithril rock | 90 tinted orange / 90 / 91 yellow boulder / 90 tinted blue |
| Depleted rock | 107 small rock |
| Cabbage / cotton / harvested plot | 79 bush / 111 white flowers / 244 |
| Farm field | 206 tilled soil |
| Bank | 92 stone house front |
| Blacksmith | 93 wooden hut front |
| Anvil | 175 anvil |
| Fletcher's hut | 125 wooden hut |
| Felled tree | 76 dead tree |
| Cottage / tower | 95 / 94 |
| River | 115 water |
| Road / bridge / square | 191 dirt / 13 planks / 108 cobblestone |
| Trees | oak 74, pine 80, maple 75 (autumn), dark oak 78; bush 79 |
| Decor | 111, 112 flowers; 119 pebbles; 109 mushroom; 118 stones; 123 rubble |

The sprite sheet has no ore-specific rocks. Copper and mithril are the iron rock drawn with a colour tint (`resources.TintedTiles`, sprite numbers past the end of the sheet).

## Architecture decisions

- **The server is authoritative for players.** Today the client moves itself and the server relays positions. That can't work for an idle game: progress has to happen while the client is closed, and inventories can't be trusted to the client. Clients send intents (`walk to tile`, `interact with object`). The server pathfinds, moves the character each tick, runs the jobs and broadcasts positions and inventories.
- **One world goroutine owns all game state.** Connection handlers send it commands over a channel instead of touching shared maps. This also fixes the current data races.
- **Pathfinding** is A* over tiles, 8-directional, using the passability rule above. A diagonal step needs both tiles beside it free, so sprites never clip a corner. Steps onto roads, bridges and village squares cost 60% of open ground, so walkers follow roads when the detour is sensible. All water (sea, rivers, lakes) blocks.
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
5. **Skills, tiers and farms.** Five skills levelling 1–99, four material tiers each, farms with cabbages and cotton, a nearest-place action window, a skills window. *(done)*
6. **Equipment, combat and health.** Five combat skills, health and food, an equipment window, four tiered camps with enemies and loot, the Fight action. Also half-speed actions and walking, a draggable camera, the route line, and a Windows build in CI. *(done; PR #14)*
7. **Crafting.** Make the gear that camps drop: swords, helmets, platebodies and shields from bars at the anvil, bows at the fletcher; weave cotton into cloth and cloth armour; cook or prepare better food. Equipment stats already exist (`resources.Equipment`).
8. **Persistence.** Save characters (skills, health, bank, equipment, action lists, found places) to disk on a Docker volume, and hash passwords. Until this lands, every deploy wipes progress, which hurts more now there's combat and gear; consider doing it before crafting.

## Decisions

- Bank accounts are private to each player.
- Iron rocks run out and respawn.
- Only one player can mine a rock at a time.
