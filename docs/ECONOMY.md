# Economy: crafting, coins and trade

The plan for the next big phase: the item quality and improving rules from `docs/CRAFTING-QUALITY.md`, made into a working economy with money, shops and player markets. Decisions made on 2026-11-21 are recorded under "Decisions"; anything else here is still a proposal.

## Decisions

- **Persistence waits.** The game is early in development; deploys wiping characters is acceptable for now, so the economy is built first.
- **Items break at 100 damage** and are gone.
- **Selling to merchants is a walk-and-sell action.** *(Changed 2026-11-21: for now, stores, banks and inns just have a Visit action; buying and selling happen in the store window, at the store.)*
- **Player markets are per village**, not one global market (see "Player markets").
- **Quality changes gear power by ±30%.**
- **Inn stays cost coins**, with the stable as a free, slower fallback.

## What the economy is for

Today everything a player gathers ends up in their own bank and stays there. An economy gives those items a purpose beyond the player's own use:

- **Specialisation.** A miner sells ore, a smith buys it and sells swords, a fighter buys swords and sells loot. Nobody has to level every skill.
- **Quality matters.** A quality-70 sword is worth far more than a quality-20 one, so improving items is a way to earn, not just a vanity goal.
- **Long-term goals.** Expensive gear and high-quality items are things to save up for over many hours.
- **Healthy flow.** Coins come into the game from loot and selling to merchants, and leave it through inn stays, repairs, purchases and market fees. If coins only came in, prices would inflate; if they only left, players would run dry. The design below balances both.

## Part 1: Crafting rules

These are the rules from `docs/CRAFTING-QUALITY.md`, with my recommended answer to each open question there.

### Quality on everything
- **Gear and tools** are individual items with **quality 1–100** and **damage 0–100**.
- **Materials** (ore, bars, logs, cloth, crops) carry quality too, as a stack's weighted average *(open question 1: option A)*. Gathered quality comes from skill level and tool quality.
- **Effective quality** = `quality × (100 − damage) / 100`.

### Tools (new)
Gathering uses a tool, carried in a new **tool slot** in the equipment window:

| Tool | Used for | Made at |
|---|---|---|
| Pickaxe | Mining | Anvil, from bars |
| Hatchet | Woodcutting | Anvil, from bars |
| Sickle | Farming | Anvil, from bars |
| Hammer | Smithing, improving metal | Anvil, from bars |
| Knife | Fletching, improving bows | Anvil, from bars |

- Each comes in the four metals. Better metal and higher quality mean faster work and better-quality gathered materials.
- Every new character gets a set of basic quality-10 copper tools, so nothing is locked behind crafting.
- Without the right tool, an action still works, but slowly (half speed) and only at quality 1 to 10.

### Making things
| Station | Skill | Makes | From |
|---|---|---|---|
| Anvil | Smithing | swords, armour, tools | bars |
| Fletcher's hut | Fletching | bows (new), arrows | logs |
| Loom (new, in each village) | Tailoring (new skill) | cloth, then cloth armour (light armour for archers and casters) | cotton |
| Staff carver (at the fletcher's hut) | Fletching | mage and necromancer staves | logs and bars |

- New item quality = `min(material quality, random around skill × 0.5)`. New items come out rough; improving is how they get good.
- Making is an action ("Make iron sword at the Oakhaven anvil"), with goals of N items or a time.

### Improving
- **"Improve [item]"** at the matching station, with goals of "until quality N", N attempts, or a time.
- Each attempt uses a little material of the item's kind, which must be better quality than the item. Success chance falls as the item's quality rises past your skill; gains shrink as quality rises; failure adds damage and the character repairs it automatically for a small quality loss.
- **The tool steps are abstract** *(open question 5)*: the activity bar shows "Hammering", "Sharpening", "Polishing" and so on as the item needs them, but they only need the station's main tool (hammer or knife), not a bag of separate items.
- Improving above your skill gives the most XP.

### Wear, repair and breakage
- Gear and tools take a little damage with use; higher quality wears slower.
- **Repair** costs coins at a village smith, about `item value × damage%`, and is one of the main coin sinks. A job repairs gear automatically on bank trips once damage passes 50%, if the player can afford it.
- **At 100 damage an item breaks and is gone** *(open question 2)*. This keeps demand for new gear in a game that runs for years. Automatic repairs make it rare for anyone paying attention, and a warning goes to chat at 75%.

### Quality and gear power *(open question 4)*
Weapon and armour bonuses scale from **70% at quality 1 to 130% at quality 100** (±30%). That makes quality clearly worth paying for without making low-quality gear useless.

### Camp drops *(open question 3)*
Camps keep dropping gear, but always at **quality 5–25**. Crafted gear can go far higher, so smiths have a market, and fighters still get something usable early.

### Rarity *(open question 6)*
Left for a later phase: it's a pure bonus on top and isn't needed for the economy to work.

## Part 2: Money

### Coins
- One currency: **coins**.
- Each player's coins live in their **bank account** as a balance, not in slots, and show at the top of the inventory window. They can't be lost on defeat.
- New characters start with **50 coins**.

### Where coins come from
| Source | How much |
|---|---|
| Camp loot | goblins 2–8, bandits 8–20, orc warriors 20–50, skeletons 50–120 per kill |
| Selling to merchants | the merchant's buying price (below) |
| Selling to players | whatever they pay, minus the market fee |
| Village orders | the order's reward (below) |

### Where coins go
| Sink | How much |
|---|---|
| Inn stays | 5 coins a rest. A character who can't pay sleeps in the stable: free, but half speed |
| Repairs | about item value × damage% |
| Buying from merchants | the merchant's selling price |
| Market fee | 5% of every sale between players |
| Equipment upkeep | none; breakage covers it |

### Item values
Every item has a **base value**. Shops and orders are priced from it, and players can use it as a guide.

| Tier | Ore / logs | Bars | Arrows (each) | Gear (weapon / body armour) |
|---|---|---|---|---|
| 1 (copper, oak) | 2 | 6 | 1 | 40 / 60 |
| 2 (iron, pine) | 5 | 15 | 2 | 120 / 180 |
| 3 (gold, maple) | 12 | 36 | 4 | 350 / 500 |
| 4 (mithril, dark oak) | 30 | 90 | 8 | 1,000 / 1,400 |

- Crops: cabbage 3, cotton 4; cloth 12.
- Gear value scales with quality: `base × (0.5 + quality / 50)`, so quality 25 is worth base, quality 75 twice base, quality 100 two and a half times base.

## Part 3: Trading

### Merchants (village general stores)
Each village gets a **general store**, run by the game.

- **Buying from the store** happens instantly from a Shop window, for any village you've found. Goods go straight to your bank. Stores sell:
  - food, at 1.5× value;
  - basic copper tools and gear at quality 15, at 1.5× value;
  - starter materials, at 2× value.

  Prices are high on purpose: players should usually be the cheaper source.
- **Selling to the store** happens in the store window, from the backpack, standing at the store (reached with a Visit action or by walking there). An earlier version had a Sell action that carried goods from the bank automatically; it may come back.
- **Prices move with supply.** Each store pays around 40% of value. Every unit it buys lowers its price for that item a little, and the price recovers over a few hours. Flooding one store with ore gets you less, so it pays to spread sales around or sell to players instead.

### Player markets (a trading post in each village)
The heart of a player economy. Each village has its own **trading post** with its own listings, so prices can differ from village to village and travel matters.

- **Listing** is an action: "List [item] at the Saltmere trading post, N at [price]". The character carries the goods from the bank to that post, where they're held safely until sold. Gear lists individually with its quality shown; materials list as stacks, labelled with their quality.
- **Buying** happens instantly from a **Market** window, for any village you've found. Browse or search that post's listings, filtered by item, tier and minimum quality, and sorted by price. The coins move straight away (to the seller's bank, minus the 5% fee), even while the seller is offline.
- **Collecting.** What you buy waits at that village's trading post until you collect it with a **Collect** action, which carries it to your bank.
- **Cancelling** a listing leaves the items at the post, ready to collect.
- **Limits:** 10 listings per player to start. A listing expires after 7 days and its items wait at the post, ready to collect.
- **Buy orders** come later: "I'll pay 20 each for 100 iron bars of quality 40 or better". Sell listings are enough to start.

### Village orders
Each village posts a few orders that refresh every few hours, such as "Wanted: 40 iron bars, quality 30+, pays 900 coins".

- Delivering one is an action, like selling, and pays more than the store would.
- Orders steer players towards things the world needs, give new players a reliable income, and put coins into the game in a controlled amount.

## What has to change in the code

| Area | Change |
|---|---|
| Persistence (later) | Deferred: deploys wipe the economy along with everything else until it's built. |
| Items | `Item{ID, Kind, Quality, Damage}` for gear and tools; stacks gain a `Quality`. Backpack, bank, equipment, messages and windows all change. |
| Values | A `Value` on every item, plus the quality multiplier. |
| Skills | Tailoring (new). |
| World | General store, loom and staff carver in each village; the anvil becomes usable. |
| Jobs | Make, improve, repair, sell and deliver actions; automatic repairs on bank trips; inn fee. |
| Combat | Gear bonuses scale with effective quality; damage from use; coin drops. |
| Market | Per-village listings and held purchases on the server; list and collect actions; market and shop windows. |
| Client | Coins in the inventory window; shop, market and orders windows; quality and damage on items. |

## Suggested order

1. **Coins and merchants.** Coin balance, item values, coin drops, the inn fee and stable, general stores (buy window, sell action). A working economy early, before the bigger item rework. *(done: stores sell food, copper gear, starter materials and oak arrows; sell trips take 3 s at the counter; demand drops 2% per unit bought, never below 20%, and recovers 3% of the gap each minute)*
2. **Item quality and tools.** Individual items, quality on stacks, tools and the tool slot, quality affecting gathering and gear. *(done, with a tool belt of one slot per tool instead of a single tool slot, so action lists don't need tool swaps. Gathered quality = 10 + 0.6 × level + tool metal bonus + 0.2 × tool quality, ±5; crafted quality is the input's, capped at 15 + 0.8 × level + tool quality / 10. Loot gear is quality 5–25, store and starting gear 15, starting tools 10. Stores pay for pieces by quality. Damage exists on items but nothing causes it until step 4.)*
3. **Making and improving.** Anvil, loom and staff recipes, Tailoring, the make and improve actions. *(Making and improving done, both by hand in the anvil or fletcher window from the backpack: gear at the anvil, bows and staves at the fletcher. New gear is quality ≈ level/2 capped by its materials. Improving is done in a window at the anvil or fletcher, on a piece from the backpack: each attempt uses a quarter of a better-quality unit of the recipe's main material and wears the tool 1%, so tools now take damage and can break. The loom and Tailoring wait: adding a building changes world generation, which moves everything on the island.)*
4. **Wear and repair.** Damage, repair costs, breakage.
5. **Player markets.** A trading post in each village; list, buy and collect.
6. **Village orders.**
7. **Rarity.**
8. **Persistence**, when development settles.

Each step fits in a phase and leaves the game playable.
