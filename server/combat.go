package main

import (
	"fmt"
	"math/rand"
	"slices"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
	"golang.org/x/image/math/f64"
)

// Combat tuning. See docs/DESIGN.md.
const (
	playerAttackTime = 9600 * time.Millisecond
	enemyAttackTime  = 12 * time.Second
	enemyRespawn     = 30 * time.Second
	enemyWanderMin   = 3 * time.Second
	enemyWanderMax   = 7 * time.Second
	enemyWanderRange = 2 // tiles from home
	enemyStep        = walkStep * 0.75
	enemyRepath      = 500 * time.Millisecond  // how often a chasing enemy re-aims at its target
	regenTime        = 1200 * time.Millisecond // per health point, out of combat
	restRegenTime    = 250 * time.Millisecond  // per health point, resting at the bank
	castRange        = 5                       // tiles, for bows and staves; swords need to be next to the enemy
	foodPerTrip      = 10
	eatBelow         = 0.5 // eat when health drops below this share of max
	retreatBelow     = 0.3 // with no food left, leave below this share
	restBelow        = 0.6 // rest at the bank before a trip below this share
	restUntil        = 0.9 // and get up again at this share
	mobIDBase        = 20000
)

// Activities shown while fighting.
const (
	activityToCamp    = "Walking to the camp"
	activitySeeking   = "Looking for an enemy"
	activityToEnemy   = "Closing in"
	activityFighting  = "Fighting"
	activityResting   = "Resting"
	activityNoArrows  = "Out of arrows"
	activityDefeated  = "Defeated"
	activityNoEnemies = "Waiting for enemies to return"
)

// Every new character starts with this much food in the bank, on top of
// their class's gear.
const startingFood = 5

// defaultClass is what a character gets if they're created without a
// choice (tests use World.login directly).
const defaultClass = resources.ClassWarrior

// mob is one enemy at a camp.
type mob struct {
	id     uint16
	kind   resources.EnemyKind
	camp   string
	home   tile
	pos    f64.Vec2
	path   []tile
	health int

	fighting        *player // the one player fighting it, if any
	attackCountdown int     // on the fighter's clock
	repathCountdown int
	wanderCountdown int // on the world clock
	dead            bool
	respawnAt       int // world tick
	dirty           bool
}

func (m *mob) tile() tile { return walkerTile(m.pos, m.path) }

func (m *mob) update() resources.ServerEntityUpdateContents {
	return resources.ServerEntityUpdateContents{
		EntityId:  m.id,
		Pos:       m.pos,
		Tile:      f64.Vec2{float64(m.kind.Sprite[0]), float64(m.kind.Sprite[1])},
		Username:  fmt.Sprintf("%s (level %d)", m.kind.Name, m.kind.Level),
		Health:    m.health,
		MaxHealth: m.kind.Health,
		Gone:      m.dead,
	}
}

func (w *World) initMobs() {
	for _, o := range w.wm.Objects {
		if o.Kind != resources.ObjectSpawn {
			continue
		}
		kind, ok := resources.FindEnemy(o.Variant)
		if !ok {
			continue
		}
		home := tile{o.X, o.Y}
		camp := ""
		if pl := w.placeAt(home, resources.PlaceCamp); pl != nil {
			camp = pl.Name
		}
		x, y := home.pixel()
		w.mobs = append(w.mobs, &mob{
			id: uint16(mobIDBase + len(w.mobs)), kind: kind, camp: camp, home: home,
			pos: f64.Vec2{x, y}, health: kind.Health, wanderCountdown: rand.Intn(ticks(enemyWanderMax)),
		})
	}
}

func (w *World) mobSnapshot() *resources.Message {
	contents := make([]resources.ServerEntityUpdateContents, len(w.mobs))
	for i, m := range w.mobs {
		contents[i] = m.update()
	}
	return resources.NewServerEntityUpdateMessage(contents)
}

// tickMobs runs enemies nobody is fighting: respawning and wandering near
// home. Fighting enemies are run from their fighter's tick, on their clock.
func (w *World) tickMobs() {
	var changed []resources.ServerEntityUpdateContents
	for _, m := range w.mobs {
		switch {
		case m.dead:
			if w.ticks >= m.respawnAt {
				m.dead = false
				m.health = m.kind.Health
				m.pos[0], m.pos[1] = m.home.pixel()
				m.path = nil
				m.dirty = true
			}
		case m.fighting == nil:
			m.wanderCountdown--
			if m.wanderCountdown <= 0 {
				m.wanderCountdown = ticks(enemyWanderMin) + rand.Intn(ticks(enemyWanderMax-enemyWanderMin))
				to := m.home.add(tile{rand.Intn(2*enemyWanderRange+1) - enemyWanderRange, rand.Intn(2*enemyWanderRange+1) - enemyWanderRange})
				if path := findPath(&w.wm, m.tile(), to); path != nil {
					m.path = path
				}
			}
			if stepAlong(&m.pos, &m.path, enemyStep) {
				m.dirty = true
			}
		}
		if m.dirty {
			changed = append(changed, m.update())
			m.dirty = false
		}
	}
	if len(changed) > 0 {
		w.broadcast(resources.NewServerEntityUpdateMessage(changed))
	}
}

// release lets an enemy go: it heals and heads home.
func (m *mob) release() {
	m.fighting = nil
	if !m.dead {
		m.health = m.kind.Health
		m.dirty = true
	}
}

// --- health ---

func (p *player) maxHealth() int {
	return resources.MaxHealth(p.level(resources.SkillHitpoints))
}

// giveStartingKit sets a new character up as their class: look, head-start
// levels, gear (equipped or banked) and some food.
func (w *World) giveStartingKit(p *player, class resources.Class) {
	p.class = class.ID
	p.sprite = f64.Vec2{float64(class.Sprite[0]), float64(class.Sprite[1])}
	p.xp[resources.SkillHitpoints] = resources.XPForLevel(resources.StartingHitpoints)
	for s, l := range class.Levels {
		p.xp[s] = resources.XPForLevel(l)
	}
	p.health = p.maxHealth()
	p.stamina = p.maxStamina()
	p.coins = resources.StartingCoins
	p.equipment = map[resources.EquipSlot]resources.Stack{}
	for _, g := range class.Gear {
		if slices.Contains(class.Equipped, g.Kind) {
			p.equipment[resources.Equipment[g.Kind].Slot] = g
		} else {
			p.bank.Add(g.Kind, g.Count)
		}
	}
	p.bank.Add(resources.ItemCabbage, startingFood)
}

func (w *World) heal(p *player, by int) {
	if h := min(p.maxHealth(), p.health+by); h != p.health {
		p.health = h
		p.healthDirty = true
		p.invDirty = true
	}
}

func (p *player) inCombat() bool {
	return p.job != nil && p.job.mob != nil && p.job.state == stateFighting
}

// tickHealth regenerates health slowly out of combat, quickly while resting.
func (w *World) tickHealth(p *player) {
	if p.health >= p.maxHealth() || p.inCombat() {
		p.regenCounter = 0
		return
	}
	every := ticks(regenTime)
	if p.job != nil && p.job.state == stateResting {
		every = ticks(restRegenTime)
	}
	p.regenCounter++
	if p.regenCounter >= every {
		p.regenCounter = 0
		w.heal(p, 1)
	}
}

func isFood(k resources.ItemKind) bool { return resources.Items[k].Heal > 0 }

// eat eats the first food in the backpack, if any.
func (w *World) eat(p *player) bool {
	for _, s := range p.backpack.Stacks {
		if isFood(s.Kind) {
			p.backpack.Remove(s.Kind, 1)
			w.heal(p, resources.Items[s.Kind].Heal)
			return true
		}
	}
	return false
}

func (p *player) carryingFood() bool {
	for _, s := range p.backpack.Stacks {
		if isFood(s.Kind) {
			return true
		}
	}
	return false
}

// --- the fight trip ---

// atBankForFight starts a fighting trip: rest if hurt, take food, head out.
func (w *World) atBankForFight(p *player) {
	if p.health < int(float64(p.maxHealth())*restBelow) {
		p.job.state = stateResting
		w.setActivity(p, activityResting)
		return
	}
	w.collectFood(p)
}

func (w *World) collectFood(p *player) {
	n := min(foodPerTrip, p.backpack.FreeSlots(), p.bank.Count(resources.ItemCabbage))
	if n > 0 {
		p.bank.Remove(resources.ItemCabbage, n)
		p.backpack.Add(resources.ItemCabbage, n)
		p.invDirty = true
	}
	pl := w.place(p.job.action.spec.Place)
	if pl == nil || !w.moveTo(p, tile{pl.X, pl.Y}) {
		w.failAction(p, activityNoPath)
		return
	}
	p.job.state = stateWalkToCamp
	w.startAction(p, activityToCamp, walkTicks(p))
}

// seekEnemy picks the nearest free enemy of the action's kind at the camp.
func (w *World) seekEnemy(p *player) {
	j := p.job
	from := p.currentTile()
	var best *mob
	for _, m := range w.mobs {
		if m.dead || m.camp != j.action.spec.Place || m.kind.ID != j.action.spec.Material ||
			(m.fighting != nil && m.fighting != p) {
			continue
		}
		if best == nil || octile(from, m.tile()) < octile(from, best.tile()) {
			best = m
		}
	}
	if best == nil {
		j.mob = nil
		j.state = stateWaitForNode
		j.countdown = ticks(rockWaitRetry)
		w.setActivity(p, activityNoEnemies)
		return
	}
	best.fighting = p
	best.path = nil // it stops wandering and squares up
	best.attackCountdown = ticks(enemyAttackTime)
	j.mob = best
	j.state = stateWalkToEnemy
	w.moveTo(p, best.tile())
	w.setActivity(p, activityToEnemy)
}

func (p *player) bonuses() resources.Bonuses { return resources.EquipmentBonuses(p.equipment) }

// inRange reports whether p can hit m from where they are. Bows and staves
// reach further than swords.
func (p *player) inRange(m *mob) bool {
	reach := 1
	if b := p.bonuses(); b.Ranged || b.Magic {
		reach = castRange
	}
	a, b := p.currentTile(), m.tile()
	return max(abs(a.x-b.x), abs(a.y-b.y)) <= reach
}

// tickFight advances a fight by one tick on the player's clock.
func (w *World) tickFight(p *player) {
	j := p.job
	m := j.mob
	if m == nil || m.dead || m.fighting != p {
		w.seekEnemy(p)
		return
	}
	switch j.state {
	case stateWalkToEnemy:
		if p.inRange(m) {
			p.path = p.path[:min(1, len(p.path))]
			if len(p.path) == 0 {
				j.state = stateFighting
				j.countdown = ticks(playerAttackTime)
				w.startAction(p, activityFighting, j.countdown)
			}
			return
		}
		if len(p.path) == 0 {
			w.moveTo(p, m.tile())
		}
		return
	case stateFighting:
	default:
		return
	}

	// The enemy closes in if it can't reach.
	if !adjacent(m.tile(), p.currentTile()) {
		m.repathCountdown--
		if m.repathCountdown <= 0 {
			m.repathCountdown = ticks(enemyRepath)
			if path := findPath(&w.wm, m.tile(), p.currentTile()); path != nil {
				m.path = path
			}
		}
		if stepAlong(&m.pos, &m.path, enemyStep) {
			m.dirty = true
		}
	}
	if !p.inRange(m) {
		j.state = stateWalkToEnemy
		w.moveTo(p, m.tile())
		return
	}

	// Look after yourself first.
	if p.health < int(float64(p.maxHealth())*eatBelow) && w.eat(p) {
		p.invDirty = true
	}
	if p.health < int(float64(p.maxHealth())*retreatBelow) && !p.carryingFood() {
		m.release()
		j.mob = nil
		w.goBank(p) // and rest there
		return
	}

	// Your swing.
	j.countdown--
	if j.countdown <= 0 {
		j.countdown = ticks(playerAttackTime)
		if w.exhausted(p) {
			return // let go and rest; the fight picks up afterwards
		}
		if !w.playerAttack(p, m) {
			return
		}
		w.spendStamina(p, staminaPerAttack)
		if m.dead {
			w.afterKill(p)
			return
		}
		w.startAction(p, activityFighting, j.countdown)
	}

	// Its swing.
	if adjacent(m.tile(), p.currentTile()) {
		m.attackCountdown--
		if m.attackCountdown <= 0 {
			m.attackCountdown = ticks(enemyAttackTime)
			w.enemyAttack(m, p)
		}
	}
}

func adjacent(a, b tile) bool { return max(abs(a.x-b.x), abs(a.y-b.y)) <= 1 }

// playerAttack resolves one attack. It reports false if the player couldn't
// attack at all (out of arrows), which ends the list.
func (w *World) playerAttack(p *player, m *mob) bool {
	b := p.bonuses()
	attackSkill, strengthSkill := b.Skills()
	if b.Ranged {
		ammo, ok := p.equipment[resources.SlotAmmo]
		if !ok || ammo.Count == 0 {
			m.release()
			w.failAction(p, activityNoArrows)
			return false
		}
		ammo.Count--
		if ammo.Count == 0 {
			delete(p.equipment, resources.SlotAmmo)
		} else {
			p.equipment[resources.SlotAmmo] = ammo
		}
		p.invDirty = true
	}
	roll := resources.AttackRoll(p.level(attackSkill), b.Accuracy)
	if rand.Float64() >= resources.HitChance(roll, resources.DefenceRoll(m.kind.Defence, 0)) {
		w.fightEvent(p, m, resources.AttackerPlayer, false, 0)
		return true // missed
	}
	dmg := min(m.health, rand.Intn(resources.MaxHit(p.level(strengthSkill), b.Strength)+1))
	m.health -= dmg
	m.dirty = true
	defer w.fightEvent(p, m, resources.AttackerPlayer, true, dmg) // after any drain healing below
	switch {
	case b.Magic:
		w.addXP(p, resources.SkillMagic, 4*dmg)
	case b.Ranged:
		w.addXP(p, resources.SkillRanged, 4*dmg)
	default:
		w.addXP(p, resources.SkillAttack, 2*dmg)
		w.addXP(p, resources.SkillStrength, 2*dmg)
	}
	w.addXP(p, resources.SkillHitpoints, dmg*4/3)
	if b.Drain > 0 {
		w.heal(p, dmg*b.Drain/100)
	}
	if m.health <= 0 {
		m.dead = true
		m.fighting = nil
		m.respawnAt = w.ticks + ticks(enemyRespawn)
	}
	return true
}

// enemyAttack resolves one enemy attack on a player.
func (w *World) enemyAttack(m *mob, p *player) {
	roll := resources.AttackRoll(m.kind.Attack, 0)
	if rand.Float64() >= resources.HitChance(roll, resources.DefenceRoll(p.level(resources.SkillDefence), p.bonuses().Defence)) {
		w.addXP(p, resources.SkillDefence, 2) // blocked
		w.fightEvent(p, m, resources.AttackerEnemy, false, 0)
		return
	}
	dmg := rand.Intn(m.kind.MaxHit + 1)
	p.health -= dmg
	p.healthDirty = true
	p.invDirty = true
	w.fightEvent(p, m, resources.AttackerEnemy, true, dmg)
	if p.health <= 0 {
		w.defeat(p, m)
	}
}

// afterKill rolls loot and decides what's next.
func (w *World) afterKill(p *player) {
	j := p.job
	kind := j.mob.kind
	j.mob = nil
	j.action.count++
	w.dropCoins(p, kind)
	for _, drop := range kind.Loot {
		if rand.Float64() < drop.Chance {
			p.backpack.Add(drop.Item, drop.Min+rand.Intn(drop.Max-drop.Min+1)) // whatever doesn't fit is left behind
		}
	}
	p.invDirty = true
	switch {
	case amountMet(p, j.action):
		w.wrapUp(p)
	case p.backpack.FreeSlots() == 0:
		w.goBank(p)
	default:
		w.seekEnemy(p)
	}
}

// defeat: the player loses what they're carrying and wakes up in their
// home village at full health. Equipment is kept. The list stops.
func (w *World) defeat(p *player, m *mob) {
	m.release()
	if p.job != nil {
		p.job.mob = nil
	}
	p.backpack.Stacks = nil
	p.path = nil
	p.pos[0], p.pos[1] = tile{w.wm.SpawnX, w.wm.SpawnY}.pixel()
	p.health = p.maxHealth()
	p.healthDirty = true
	w.failAction(p, activityDefeated)
	if p.sess != nil {
		p.sess.push(resources.NewChatMessage(p.id, fmt.Sprintf(
			"You were defeated by a %s and lost what you were carrying. You wake up back in the village.", m.kind.Name)))
	}
}

// --- equipment ---

// equipCommand equips an item from the bank or puts one back.
func (w *World) equipCommand(p *player, c resources.EquipCommandContents) {
	defer func() { p.invDirty = true }()
	switch c.Op {
	case resources.EquipOpUnequip:
		w.unequip(p, c.Slot)
	case resources.EquipOpEquip:
		def, ok := resources.Equipment[c.Item]
		if !ok || p.bank.Count(c.Item) == 0 || p.level(def.Skill) < def.Level {
			return
		}
		// Two hands or a shield, not both.
		if def.TwoHanded && !w.unequip(p, resources.SlotShield) {
			return
		}
		if def.Slot == resources.SlotShield {
			if weapon, ok := p.equipment[resources.SlotWeapon]; ok && resources.Equipment[weapon.Kind].TwoHanded && !w.unequip(p, resources.SlotWeapon) {
				return
			}
		}
		count := 1
		if def.Slot == resources.SlotAmmo {
			count = p.bank.Count(c.Item) // all of that kind
			if cur, ok := p.equipment[resources.SlotAmmo]; ok && cur.Kind == c.Item {
				p.bank.Remove(c.Item, count)
				cur.Count += count
				p.equipment[resources.SlotAmmo] = cur
				return
			}
		}
		p.bank.Remove(c.Item, count)
		if !w.unequip(p, def.Slot) {
			p.bank.Add(c.Item, count) // no room to put the old one away
			return
		}
		p.equipment[def.Slot] = resources.Stack{Kind: c.Item, Count: count}
	}
}

// unequip puts a slot's item back in the bank, if there's room. An empty
// slot counts as done.
func (w *World) unequip(p *player, slot resources.EquipSlot) bool {
	cur, ok := p.equipment[slot]
	if !ok {
		return true
	}
	if added := p.bank.Add(cur.Kind, cur.Count); added < cur.Count {
		p.bank.Remove(cur.Kind, added)
		return false
	}
	delete(p.equipment, slot)
	return true
}

func (p *player) equippedItems() []resources.EquippedItem {
	var out []resources.EquippedItem
	for _, slot := range resources.EquipSlots {
		if s, ok := p.equipment[slot]; ok {
			out = append(out, resources.EquippedItem{Slot: slot, Item: s.Kind, Count: s.Count})
		}
	}
	return out
}

// --- the fight panel ---

func (w *World) fightContents(p *player, m *mob) resources.FightContents {
	return resources.FightContents{
		Active:          true,
		EnemyName:       fmt.Sprintf("%s (level %d)", m.kind.Name, m.kind.Level),
		EnemySprite:     m.kind.Sprite,
		EnemyHealth:     max(m.health, 0),
		EnemyMaxHealth:  m.kind.Health,
		PlayerHealth:    max(p.health, 0),
		PlayerMaxHealth: p.maxHealth(),
	}
}

// fightEvent tells the fighter about one attack, for the fight panel's
// animation.
func (w *World) fightEvent(p *player, m *mob, attacker string, hit bool, dmg int) {
	if p.sess == nil {
		return
	}
	c := w.fightContents(p, m)
	c.Attacker, c.Hit, c.Damage = attacker, hit, dmg
	p.sess.push(resources.NewFightMessage(c))
}

// sendFightState tells the player when they start fighting an enemy or
// stop, so the fight panel opens and closes.
func (w *World) sendFightState(p *player) {
	var cur *mob
	if p.job != nil && p.job.state == stateFighting {
		cur = p.job.mob
	}
	if cur == p.fightShown || p.sess == nil {
		p.fightShown = cur
		return
	}
	p.fightShown = cur
	if cur == nil {
		p.sess.push(resources.NewFightMessage(resources.FightContents{Active: false}))
		return
	}
	p.sess.push(resources.NewFightMessage(w.fightContents(p, cur)))
}
