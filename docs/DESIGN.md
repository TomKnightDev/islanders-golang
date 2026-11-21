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
- **Adding actions.** The Actions window holds just the list and its buttons: **Add action**, **Start**/**Stop**. **Add action** opens a separate picker window, which stays open while you add several. It shows one row per action (Mine, Chop, Farm, Smelt, Fletch), at the **nearest found place** for it, with its distance and a button per material. Only what the player can do now is shown: actions with no place found yet, and materials above the player's level, are left out.
- **The action list** holds up to 5 actions. Clicking an action adds it to the end. Adding one to an empty list starts the list straight away. Entries can be moved up or down or removed.
- **Goals.** Each action has an amount: gather, make or defeat **10, 20, 50 or 100**, picked with buttons and changeable at any time. New actions start at 10. Only what the action itself does counts, not what's already in the bank.
- **Running the list.** **Start** works through the list from the top. When an action's goal is met, the character first banks anything it's carrying, then moves on. When the list is empty the character goes idle.
- **Stop** halts the list where it is; anything carried stays in the backpack. Reordering or removing the running action stops it and starts whatever is now first.
- **Clicks never cancel the list.** While the list is running, clicking the map does nothing. While it's stopped, clicking walks, which is how the player explores and finds new places.
- **Failures.** If the bank account is full, or there's no way to a place, the list stops and the action stays in it, so the player can sort things out and press Start again.

### Skills
The game is meant to be played for hundreds of hours, so skills level slowly, with lots to unlock on the way.

- **Gathering and crafting skills:** Mining, Woodcutting, Farming, Smithing (smelting), Fletching. Each action trains one. The six combat skills are under Combat, and Stamina under Stamina and inns.
- **Repair** trains by repairing worn gear and tools (see Repairing).
- **Levels 1–99.** The XP curve roughly doubles every seven levels: level 2 at 83 XP, level 50 at about 100,000, level 99 at about 13 million.
- **XP** comes from every item gathered, every input crafted and every enemy defeated. Higher tiers give more. Every award is then halved (`xpRateNum`/`xpRateDen` in `server/jobs.go`), so the tables below show XP before that.
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
| Base speed | 1 ore per 8 s | 1 log per 12 s | 1 crop per 10 s |
| A node holds | 5–10 ore | 1–4 logs | 2–5 crops |
| When empty | depleted rock, 30 s | felled tree, 60 s | harvested plot, 45 s |

Mines are ranked by distance from the spawn: the nearest has copper and iron, the next iron and gold, the farthest gold and mithril. Forests have whichever trees grew there. Every farm has rows of cabbages and cotton. Only trees inside a named forest can be chopped; scattered trees are scenery. A node with no walkable tile beside it is skipped.

### Visiting buildings
- **Visit** is an action for a village's **bank**, **store**, **inn**, **anvil** or **fletcher**. The character walks there, the action finishes, and the list stops so the building's window opens; **Start** carries on with the rest of the list. Walking up to a building by clicking does the same.
- **Store:** buy into the backpack and sell from it. Prices are in `docs/ECONOMY.md`.
- **Inn:** a Quests and Rumours window. There are no quests yet. Rumours give directions from the inn to the three nearest places the player hasn't found ("Travellers speak of a mine called Greyrock, about 40 tiles north-east of here, where copper ore is mined.").

### The bank
- Each player has their own bank account, shared between every village's bank. Its contents are only visible at a bank: walk up to one with the action list stopped and the **Bank** window opens, showing the backpack beside the bank.
- In the Bank window, items move either way (gear and tools piece by piece, materials one at a time or all at once, or **Deposit all**), and gear in the bank can be equipped. Elsewhere, gear is equipped from the backpack.
- Away from a bank, everything a character collects or buys goes into the backpack. Jobs still bank on their own as they work.

### Mining
- Mines only at the action's mine. The character walks to a free rock there and mines one **iron ore** per swing.
- When the backpack is full, or the goal is met, the character walks to the nearest bank, deposits, and goes back to the same rock.
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
| Makes | 1 bar from 1 ore, 12 s | 10 arrows from 1 log, 8 s |
| Ends early | "No ore to smelt" | "No logs to fletch" |

### Combat
- **Six combat skills**: Attack (melee accuracy), Strength (melee max hit), Defence (how often enemies hit you), Ranged (bow accuracy and max hit), Magic (staff accuracy and max hit) and Hitpoints (max health). Hitpoints starts at level 10; max health is 10 × the Hitpoints level, so 100 to start.
- **XP.** Each point of melee damage earns 2 Melee and 2 Strength XP; ranged damage earns 4 Ranged XP; all damage earns 1⅓ Hitpoints XP; each enemy attack you block earns 2 Defence XP; each attack's stamina earns Stamina XP. It's all paid **only when the enemy dies**: a fight that ends any other way (resting, defeat, stopping) pays nothing, so nobody can train on an enemy they can't beat.
- **Formulas**, scaled 10× to match the health numbers: an attack roll `(level + 8) × (accuracy + 64)` against a defence roll `(level + 9) × (armour + 64)` sets the hit chance, and `(level + 8) × (strength + 64) / 64` is the max hit.
- **Health** regenerates slowly out of combat and quickly while resting at a bank. **Food** heals when eaten: a cabbage restores 30.
- **Defeat.** At 0 health you lose everything in the backpack (not your equipment), wake up in your spawn village at full health, and the action list stops.

### Stamina and inns
- **Stamina** is used up by work: each gathering swing, smelt or fletch costs 1, each attack costs 2. Walking and idling cost nothing.
- **Running out.** At the start of the next step, a character with no stamina lets go of their rock, tree or enemy and walks to the **nearest inn**. They relax there, recovering 2 stamina a second, until fully rested, then go back to the same action; progress such as kills carries over.
- **Inns** are the only place stamina comes back. Every village has one (the open-doored wooden building).
- **The Stamina skill** trains like any other: each point spent gives 1 XP. Max stamina is 100 at level 1 and rises by 5 a level, to 590 at level 99. At level 1 that's about 13 minutes of mining or 8 of fighting before a rest.
- Stamina shows under health in the Info window.

### Coins and stores
Every player has a coin balance in their bank account (50 to start; never lost on defeat), shown in the inventory window. Coins come from enemy drops and selling to village general stores; they go on inn stays (5 coins, or a free half-speed rest in the stable) and buying from stores. Each village has a general store (the gold-door building): visit it to buy into the backpack or sell from it. Full details and what comes next are in `docs/ECONOMY.md`.

### Classes
A new name gets a **Create your character** screen before the character exists. Each class sets the character's look, starting gear and a head start; after that anyone can train anything.

| Class | Look (characters.png) | Starts with | Head start |
|---|---|---|---|
| Warrior | grey knight (8, 0) | copper sword and shield | Melee 5, Strength 5, Defence 3 |
| Archer | elf (0, 24) | oak shortbow, 150 oak arrows | Ranged 5, Defence 3 |
| Mage | lavender wizard (16, 0) | apprentice staff | Magic 5, Defence 3 |
| Necromancer | necromancer (0, 16) | bone staff | Magic 5, Defence 3 |

### Magic
- **Staves** are two-handed and cast from up to 5 tiles with no ammo, using Magic for accuracy and max hit; each point of damage gives 4 Magic XP.
- **Mage staves** (apprentice, adept, battle, archmage; Magic 1/15/30/45) hit hardest.
- **Necromancer staves** (bone, grave, wraith, lich; Magic 1/15/30/45) hit softer but heal the wielder for 25% of the damage dealt.
- Staves drop from camps at their tier, like other gear. There are no spells, runes or spellbook yet.

### Equipment
- **Quality.** Gear, tools and materials have a quality from 1 to 100 (see `docs/ECONOMY.md`). Gear and tools are individual pieces, each with its own quality. A slot holds one piece, or a stack of one material at one quality: new items only join a stack of the same quality, and **Combine selected** (tick stacks of one material in the inventory or bank window) merges just those stacks at their average quality, so poor material can be combined without touching the good. Gear bonuses scale from 70% at quality 0 to 130% at 100.
- **Tool belt.** A pickaxe, hatchet, sickle, hammer and knife each have their own slot. A tool speeds up its action (up to 30% faster for mithril at quality 100) and raises the quality of what it gathers or makes; without one, the action runs at half speed and gathers only quality 1–10. Everyone starts with copper tools at quality 10; stores sell them too.
- **Slots:** weapon, ammo, head, body, shield. The Equipment window shows them, your combat stats, and the gear in your bank; click to equip, Remove to put back. Gear above your level is greyed out with the level it needs.
- **Weapons:** a training sword; copper, iron, gold and mithril swords (Melee 1/15/30/45); oak, pine, maple and dark oak shortbows (Ranged 1/15/30/45). Bows are two-handed, attack from up to 5 tiles, and use one arrow from the ammo slot per shot; equipping arrows moves all of that kind from the bank. Better arrows hit harder.
- **Armour:** helmet, platebody and shield in each metal (Defence 1/15/30/45).
- **Starting kit:** set by the character's class (see Classes), plus 5 cabbages in the bank.
- Gear comes from camp loot, stores, or making it (see Making gear).

### Camps and fighting
- **Four camps**, easiest nearest the spawn: Goblin Camp (goblins, level 3), Bandit Hideout (bandits, 12), Orc Warband (orc warriors, 28), Skeleton Ruins (skeletons, 45). Each is a dirt clearing with huts (ruins for the skeletons) and four enemies who wander near home, with health bars when hurt.
- **Fight** is an action at a found camp, with a goal of N kills. Each trip: deposit at the bank (resting first below 60% health until 90%), take up to 10 food, walk to the camp, then fight enemy after enemy. One player fights an enemy at a time; it respawns 30 s after it dies.
- **In a fight** the character eats below half health, retreats to the bank to rest when below 30% with no food left, and banks the loot when the backpack is full. A bow user runs out of arrows and the list stops.
- **Loot** per kill rolls each drop in the enemy's table: their tier's ore and bars, food, arrows, and small chances of each piece of their tier's gear.
- Fights run on the fighter's clock.
- **The fight panel** sits above the action bar while fighting: your character and the enemy face each other with their names and health bars, "VS" between. Whoever attacks lunges towards the other; a hit flashes the target red and floats the damage up (gold for damage you deal, red for damage you take), a miss floats "Miss". After the last blow the panel stays a moment, with a defeated enemy greyed out.

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

### Making gear
- Gear is made by hand at the station, not from the action list: **Visit** the village's **anvil** or **fletcher** (or walk up to it) and use the **Make** tab of its window. It lists the recipes made there that the player has the level for, a row per metal or wood; picking one shows the ingredients needed against what's in the backpack, and the tool. **Make 1**, **Make 5** or **Make all** then works through them, the pieces going into the backpack. Making stops when done, out of materials, too tired, or on walking away.
- **Anvil** (Smithing, hammer), from each metal's bars: the five tools (1 bar), helmets (2), swords (2), shields (3) and bodies (4). Copper needs level 1 plus 0–8 depending on the piece; iron, gold and mithril start at 15, 30 and 45.
- **Fletcher** (Fletching, knife): shortbows (3 logs) and staves (2 logs and a bar of the matching tier: oak with copper, pine with iron, maple with gold, dark oak with mithril), at the wood's level plus 4 (bows) or 8 (staves). Arrows are still a Fletch action.
- A piece takes 10 s per bar at the anvil, or 8 s per log or bar at the fletcher, scaled by skill and tool; without the tool it's half speed.
- **New gear comes out rough:** quality around half the maker's level (plus a tenth of their tool's quality; halved again with no tool), and never better than the ingredients.

### Repairing
- Any worn piece (wear shows as "40% worn") can be repaired anywhere with its **Repair** button: in the Equipment window for what's worn, the Inventory window for the backpack, and the Bank window at a bank.
- Repairing takes the wear back to 0% but costs quality: about 1 for every 20% of wear at Repair level 1, down to a third of that at level 99, and always at least 1.
- Each repair gives Repair XP: half a point per 1% of wear mended (before the XP rate), so it rises slowly.

### Improving
- Improving also happens at the station: the **Improve** tab of the **anvil** (metal gear and tools) or **fletcher** (bows and staves) window lists the pieces in the backpack that can be improved there.
- Picking a piece shows what it needs: the recipe's level, units of its main material (bars for metal, logs for bows and staves) **better than the piece's quality**, carried in the backpack, and the matching tool on the belt (hammer or knife). **Start improving** then works on it, an attempt every 10 s (scaled by skill and tool).
- **Each attempt** uses a quarter of a unit (the cheapest one still better than the piece) and wears the tool by 1%; at 100% the tool breaks and is gone. Success raises the piece's quality by 1 or more, never past the material's quality. The chance is 85% up to the player's level and falls 2% for each quality point above it (never below 10%). A failure costs 1 quality.
- Each attempt gives the recipe's XP per unit of ingredient, plus 5% for each quality point the piece is above the player's level.
- Improving stops on **Stop**, on walking away, when no carried material is better than the piece, when the tool breaks, or when the player is too tired.

### Screen layout
The **Guide** (the Info window's Guide button, or F1) explains the game page by page: getting started, the action list, skills, stamina, banks and stores, quality and tools, making gear (with every recipe and its level), improving and fighting. A **bar along the top** has a link for each window (Info, Skills, Actions, Inventory, Equipment, Chat, Guide) that opens or closes it, and **Log out**, which goes back to the login screen; the character stays in the world. Each time the game loads, the windows snap to their places: **Info** top left with **Skills** below it, **Actions** top right (the action picker opens beside it), **Inventory** bottom right, and **Chat** bottom left with its message box along the bottom (Enter sends). The progress bar sits bottom middle. Windows can be moved during a session.

### Camera and route
The camera follows the character. Click and drag the map to look around; a plain click still walks there. While the camera is free, a **Follow my character** button at the top of the screen brings it back. Your own route is drawn as a line from your character to where they're heading; nobody else sees it.

### Hover information
Pointing at a building, rock or tree on the map shows a tooltip with its name and what it's for, including where it is ("Saltmere's bank", "Iron rock at Old Quarry") and, for trees, whether they can be chopped. Items in the inventory window have tooltips too. The inventory window doesn't show the current activity; the progress bar does.

### Activity and progress bar
The character shows what it's doing ("Mining", "Smelting", "Walking to the bank", "Waiting for a free rock", "Bank full") in a bar at the bottom middle of the screen. The bar fills green over each timed step: every mining swing, every smelt, every walk. Open-ended states show an empty bar, and the bar hides when idle.

### Idle behaviour
The server owns the character and simulates it, so loops keep running at full speed while the player is disconnected. Logging back in shows whatever was earned in the meantime. Offline characters stay visible to other players while they work. This works because characters already stay in server memory after disconnecting. Once persistence lands, progress survives server restarts too.

## Starting numbers

These are tuning values, kept in one place in the server code.

| Value | Start |
|---|---|
| Walk speed | 1.5 tiles/s |
| XP rate | ½ of the listed XP values, applied to every award |
| Mining swing | 8 s per ore |
| Ore per rock | 5–10, random |
| Rock respawn | 30 s |
| Backpack size | 20 slots |
| Bank account size | 100 slots |
| Smelting | 1 ore → 1 bar, 12 s |
| Chopping | 1 log per 12 s; trees hold 1–4, regrow in 60 s |
| Fletching | 1 log → 10 arrows, 8 s |
| Farming | 1 crop per 10 s; plants hold 2–5, regrow in 45 s |
| Your attack / enemy attack | every 9.6 s / every 12 s |
| Enemy respawn | 30 s |
| Health regeneration | 1 per 1.2 s, 1 per 0.25 s resting at the bank |
| Server tick | 25 ms |

## World

The island is generated by `server/worldgen.go` from a fixed seed, so it's the same on every start and deploy and a shared location always points to the same spot. It's 160×160 tiles.

- **Island.** Noise shaped by a falloff from the centre gives an irregular coastline: deep sea, shallow water, a sandy beach, then grassland. The map edge is always open sea.
- **Rivers** start in the highlands and wander downhill to the sea. They're one tile wide up high and two wide lower down, and where they meander over flat ground they spread into small lakes. Rivers can't be walked through.
- **Villages** (4): a cobbled square with the bank, blacksmith and anvil on its north side, a fletcher's hut on its south side, an inn, a general store, and cottages around it. The village nearest the middle, Oakhaven, has a tower and is where everyone spawns.
- **Mines** (3): a patch of rubble-strewn ground with six rocks, on higher ground away from villages.
- **Roads** link the villages (a minimum spanning tree) and run from each mine to its nearest village. They prefer open ground, merge where they meet, clear trees in their way, and cross rivers on wooden bridges.
- **Farms** (3): a tilled field a short walk from a village, with rows of cabbages and cotton, each row beside a walkable one. Roads lead to them and never cut through the crops.
- **Named forests** (4): the largest woods get a name (Whisperwood, Elderglade, Mossy Hollow, Thornwood) and become places where trees can be chopped.
- **Forests** grow where a second noise field is high: dense woods on a darker forest floor with mushrooms, thinning out at the edges, with some autumn-coloured patches. Elsewhere there are scattered trees, bushes and flower meadows.
- **Reachability.** After generation, every bank, blacksmith, anvil and rock is checked to be walkable from the spawn; anything cut off would get a road.

### Layers
| Layer | Holds | Blocks movement |
|---|---|---|
| 0 Ground | grass, sand, water, roads, bridges, cobbles, forest floor | water only |
| 1 Objects | trees, bushes, buildings, rocks | always |
| 2 Decor | flowers, mushrooms, rubble | never |

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
| Inn | 210 open wooden door |
| General store | 215 gold door |
| Felled tree | 76 dead tree |
| Cottage / tower | 95 / 94 |
| River | 115 water |
| Road / bridge / square | 191 dirt / 13 planks / 108 cobblestone |
| Trees | oak 74, pine 80, maple 75 (autumn), dark oak 78; bush 79 |
| Decor | 111, 112 flowers; 109 mushroom; 123 rubble |
| Weather (not used yet) | 118 / 134 rain, 119 / 135 snow (two animation frames each) |

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
4. **Smelting and the action window.** Found places, a 5-action list with amount or time goals, smelting trips. *(done; time goals were later removed in favour of fixed amounts)*
5. **Skills, tiers and farms.** Five skills levelling 1–99, four material tiers each, farms with cabbages and cotton, a nearest-place action window, a skills window. *(done)*
6. **Equipment, combat and health.** Five combat skills, health and food, an equipment window, four tiered camps with enemies and loot, the Fight action. Also half-speed actions and walking, a draggable camera, the route line, and a Windows build in CI. *(done)*
6½. **Classes and magic.** A character creation screen with four classes (warrior, archer, mage, necromancer), the Magic skill and mage and necromancer staves. Also full speed while offline, doubled action timers, and stamina with inns in every village. *(done)*
7. **Crafting and the economy.** Built in the order set out in `docs/ECONOMY.md`: coins and merchants, item quality and tools, making and improving, wear and repair, per-village player markets, village orders, rarity.
   Original crafting notes: Make the gear that camps drop: swords, helmets, platebodies and shields from bars at the anvil, bows at the fletcher; weave cotton into cloth and cloth armour; cook or prepare better food. Equipment stats already exist (`resources.Equipment`). An item quality and improving system is proposed in `docs/CRAFTING-QUALITY.md`. The economy built on it (coins, merchants, a player market, village orders) is proposed in `docs/ECONOMY.md`, with a suggested order.
8. **Persistence.** Save characters (skills, health, bank, equipment, action lists, found places, coins, listings) to disk on a Docker volume, and hash passwords. Deferred until development settles (decided 2026-11-21); until then every deploy wipes progress.

## Decisions

- Bank accounts are private to each player.
- Iron rocks run out and respawn.
- Only one player can mine a rock at a time.
