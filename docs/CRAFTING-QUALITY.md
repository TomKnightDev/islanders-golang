# Item quality and improving

A proposal for a crafting system where items have a quality level and players improve them over time. Nothing here is built yet. Read with `docs/DESIGN.md`; this would replace or extend phase 7 (Crafting).

## The idea

Every crafted item has a **quality** that the player can raise, a little at a time, by working on it. A sword made today is rough; with skill, good materials and many hours of work it becomes something worth showing off. Quality is the long-term goal for crafters, the way levels are for everyone.

Islanders is semi-idle, so improving is an **action** like mining: plan it, and the character works on the item until a goal is met.

## Quality

- Crafted gear (swords, bows, armour) and new **tools** (pickaxe, hatchet, sickle, hammer) have a **quality from 1 to 100** and **damage from 0 to 100**. Each piece is an individual item, not a stack.
- Raw materials and bars have quality too. Two options:
  - **A. Stack quality.** A stack keeps a weighted-average quality, so adding 10 ore at quality 30 to 10 at quality 20 gives 20 at quality 25. Simple, and it keeps bank slots meaningful.
  - **B. No quality on materials.** Only gear and tools have quality. Simpler, but loses the "better material, better result" rule below.
  - Recommendation: **A**. Gathered quality is set by skill level and tool quality.
- **Effective quality** accounts for damage: `quality × (100 − damage) / 100`. A badly damaged sword works like a poor one.

### What quality does
| Item | Effect of quality |
|---|---|
| Weapons | Accuracy and strength bonuses scale from about 50% at quality 1 to 150% at quality 100 |
| Armour | Defence bonus scales the same way |
| Tools | Faster actions (for example up to 25% faster at quality 100) and higher quality of what they gather |
| Materials | Cap the quality of what's made or improved with them |
| Everything | Less wear per use at higher quality |

## Creating

- Smithing makes gear at the **anvil** from bars, Fletching makes bows at the fletcher, and a new **Tailoring** skill turns cotton into cloth and cloth armour at a loom.
- New item quality = `min(material quality, random around skill × 0.5)` *(tunable)*. New items come out low, and improving is how they get good.

## Improving

- **"Improve [item]"** at a station, for example "Improve iron sword at the Oakhaven anvil".
- **Goals:** "until quality N" (new), "for N minutes", or "N attempts".
- **Each attempt:**
  - **Step.** The item needs a particular kind of work next, which changes after every attempt. For metal: **hammer** (flatten dents), **whetstone** (sharpen), **water** (temper), **pelt** (polish), or **more metal** (uses a little of the bar). The tools must be in the backpack, and the character uses whichever the item needs. Tools matter without the player clicking through each step.
  - **Success chance** comes from skill against the item's quality, nudged by tool quality. It's high while the item is below your skill and falls steeply above it.
  - **Gain** is about `(skill − quality) / 20`, clamped to a small minimum, so it shrinks as quality rises. Getting the last few points is slow by design.
  - **Failure** adds damage. The character repairs automatically, losing about 0.5% quality, then carries on.
  - **Material** must be of higher quality than the item, or improving stops with "Your materials aren't good enough".
- **Caps:** your skill and the material's quality set how far an item can go. Pushing well past your skill is possible but slow.
- **XP:** much higher when the item is above your skill, so improving is the main way to level crafting skills late on.

## Rarity

- Each crafting or improving attempt has a small chance (for example 1 in 2,000) of a **moment of inspiration**, which makes the item **Rare**. Later lucky rolls can make it **Supreme**, then **Fantastic**.
- Bonuses by tier:
  - **Tools:** gather faster, gather better quality, and wear more slowly.
  - **Weapons:** a chance of critical hits.
  - **Armour:** extra damage reduction.

## Wear, repair and breakage

- Using gear and tools adds a little damage: each swing, chop or hit, and each hit taken for armour. Higher quality wears more slowly.
- **Repair** at a station (or automatically during the action list) clears damage for a small quality loss and a little material.
- At 100 damage the item **breaks** and is gone. This keeps demand for crafted gear in a long-running game.
- Decay over time is left out: it doesn't suit an idle game where players are away a lot.

## What changes in the code

- **Item instances.** Inventories are slots of `Stack{Kind, Count}` today. Gear needs individual items: `Item{ID, Kind, Quality, Damage, Rarity}`. Stacks of materials get a `Quality` field (option A). Backpack, bank, equipment, messages and the Inventory and Equipment windows all change.
- **Combat formulas** take effective quality into account in `EquipmentBonuses`.
- **Gathering** takes tool quality and skill into account for speed and the quality of what's gathered.
- **New:** tools, a Tailoring skill and loom, anvil recipes, improve and repair jobs, and a rarity roll.
- **Persistence first.** Improving an item over hours or days is only worth it if it survives a deploy. Today every deploy wipes characters, so persistence (phase 8) should land before this.

## Questions to decide

1. Quality on materials (option A) or only on gear and tools (option B)?
2. Should items break at 100 damage, or stay unusable until repaired?
3. Should crafted gear replace camp drops, or should camps keep dropping low-quality gear?
4. How much should quality matter? A ±50% swing on bonuses makes it central; ±20% makes it a nice-to-have.
5. Should the improving steps need real tools to carry (hammer, whetstone, water, pelt), or be an abstract "improve" step?
6. Rarity: include now, or later?

## Suggested order

1. **Persistence** (phase 8): characters, banks and equipment saved across deploys.
2. **Item instances and quality**: the data model change, with existing gear given a starting quality. Effects on combat and tools.
3. **Creating gear**: anvil, fletcher and loom recipes; tools; Tailoring.
4. **Improving and repair**: as actions, with the step rotation, damage and goals.
5. **Wear and rarity.**
