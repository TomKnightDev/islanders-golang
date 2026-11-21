package gui

import (
	"fmt"
	"strings"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// Guide is the player guide: a page list on the left and the chosen page on
// the right. It opens from the Info window or with F1.
type Guide struct {
	mgr  *renderer.Manager
	Open bool
	page int
}

func NewGuide(mgr *renderer.Manager) *Guide { return &Guide{mgr: mgr} }

// guidePage is one page: a title and a function that draws it.
type guidePage struct {
	title string
	draw  func()
}

var guidePages = []guidePage{
	{"Getting started", guideStart},
	{"The action list", guideActions},
	{"Skills and XP", guideSkills},
	{"Stamina and inns", guideStamina},
	{"Banks and stores", guideBanks},
	{"Quality and tools", guideQuality},
	{"Making gear", guideMaking},
	{"Improving gear", guideImproving},
	{"Repairing", guideRepairing},
	{"Fighting", guideFighting},
	{"Hunting", guideHunting},
}

const guideWrap = 620

func (g *Guide) Update() error {
	if !g.Open {
		return nil
	}
	w, h := screenSize()
	imgui.SetNextWindowPosV(imgui.Vec2{X: w / 2, Y: h / 2}, imgui.ConditionAppearing, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: 960, Y: 620}, imgui.ConditionAppearing)
	imgui.BeginV("Guide", &g.Open, 0)

	imgui.BeginChildV("pages", imgui.Vec2{X: 250, Y: 0}, true, 0)
	for i, p := range guidePages {
		if imgui.SelectableV(p.title, i == g.page, 0, imgui.Vec2{}) {
			g.page = i
		}
	}
	imgui.EndChild()

	imgui.SameLine()
	imgui.BeginChildV("page", imgui.Vec2{}, false, 0)
	p := guidePages[g.page]
	imgui.Text(p.title)
	imgui.Separator()
	imgui.PushTextWrapPosV(guideWrap)
	p.draw()
	imgui.PopTextWrapPos()
	imgui.EndChild()

	imgui.End()
	return nil
}

func (g *Guide) Draw(screen *ebiten.Image) {}

// para draws a paragraph with space after it.
func para(text string) {
	imgui.Text(text)
	imgui.Spacing()
}

// point draws a bullet point.
func point(text string) {
	imgui.Bullet()
	imgui.SameLine()
	imgui.Text(text)
}

func heading(text string) {
	imgui.Spacing()
	imgui.Text(text)
}

func guideStart() {
	para("Islanders is played mostly by planning. Explore the island to find villages, mines, forests, farms and camps, then give your character a list of things to do. They work through it on their own, even while you're offline.")
	point("Click the map to walk. Walking into a place finds it, and found places appear in the action picker.")
	point("Drag the map to look around; scroll to zoom. The camera follows you again when you walk.")
	point("The bar along the top opens and closes each window, and has Log out. Your character stays in the world, still working, while you're logged out.")
	point("Every village has a bank, a store, an inn, a blacksmith, an anvil, a fletcher and a tailor.")
	point("Walk up to a building, or visit it from the action list, to open its window.")
	point("Inns tell rumours of places you haven't found yet, with directions.")
	heading("A first loop")
	para("Mine copper at a mine. Your character banks the ore when the backpack is full. Smelt it into bars at a village blacksmith, take the bars out at the bank, and make a copper tool at the anvil.")
}

func guideActions() {
	para(fmt.Sprintf("The action list holds up to %d actions. Open the picker with Add action; it only shows what you can do now, at the nearest place you've found for each.", resources.MaxActions))
	point("Mine, Chop and Farm gather from rocks, trees and crops, banking whenever the backpack fills. Hunt stalks deer in a forest for leather and venison.")
	point("Smelt turns ore from the bank into bars, Fletch turns logs into arrows, and Weave turns cotton into cloth. Each carries the input from the bank to the station and back.")
	point("Fight takes food from the bank and fights enemies at a camp.")
	point("Visit walks to a village's bank, store, inn, anvil, fletcher or tailor, then stops the list so its window opens.")
	heading("Goals")
	para(fmt.Sprintf("Each action has a goal of %s: that many gathered, made or defeated by the action itself. What's already in the bank doesn't count.", joinInts(resources.GoalAmounts)))
	point("Adding an action to an empty list starts it straight away.")
	point("Stop pauses the list. While it's stopped, clicking the map walks; while it runs, clicks are ignored.")
	point("When an action's goal is met, the character banks what's carried and moves on to the next.")
}

func guideSkills() {
	para("There are fifteen skills, each from level 1 to 99: Melee, Strength, Defence, Ranged, Magic and Hitpoints for fighting; Mining, Woodcutting, Farming, Hunting, Smithing, Fletching and Tailoring for gathering and making; Stamina; and Repair.")
	point("XP comes at the end of each step: each item gathered, each bar or arrow made, each piece of gear made or improving attempt.")
	point("Fight XP is held until the enemy dies. A fight that ends any other way pays nothing.")
	point("Each level makes that skill's work a little faster (about 40% faster by level 99) and unlocks better materials and recipes.")
	point("The Skills window shows your levels and progress to the next one.")
}

func guideStamina() {
	para("Work tires your character out. Each step of work uses 1 stamina, and each attack 2.")
	point("At zero, your character lets go of what they were doing and walks to the nearest inn to rest, then carries on.")
	point(fmt.Sprintf("Resting at an inn costs %d coins. Without the coins, your character sleeps in the stable for free, but it takes twice as long.", resources.InnFee))
	point("Stamina trains like any other skill: each point used is 1 XP. Each level adds 5 to your maximum.")
}

func guideBanks() {
	heading("The bank")
	para("Your bank account is shared between every village's bank, but you can only see it standing at one, with the action list stopped.")
	point("The Bank window moves items between your backpack and the bank, and equips gear from the bank.")
	point("Away from a bank, everything you collect or buy goes into your backpack. Jobs still bank on their own as they work.")
	point("Defeat in a fight loses everything in your backpack, so bank what matters.")
	heading("Stores")
	para("Each village store sells food, copper gear and tools, and starter materials, into your backpack. It buys almost anything from your backpack.")
	point(fmt.Sprintf("A store pays about %d%% of an item's value; better gear and tools pay more.", int(resources.StoreBuysAt*100)))
	point("Every item a store buys lowers its price for that kind a little; prices recover over a couple of hours. Spread your sales around.")
	point(fmt.Sprintf("You start with %d coins. Coins also drop from enemies.", resources.StartingCoins))
}

func guideQuality() {
	para("Gear, tools and materials have a quality from 1 to 100. Each piece of gear or tool has its own quality. Materials stack only with the same quality, so each line in your backpack or bank is one slot: \"Copper ore x3 q23\".")
	point("To combine stacks, tick the ones you want (they must be the same material), then press Combine selected. They merge into as few slots as possible at their average quality, so you can combine just your poor ore and keep the good ore apart.")
	point("Better gear is stronger: its bonuses range from 70% at quality 0 to 130% at quality 100.")
	point("Wear lowers a piece's working quality. A tool at 100% wear breaks and is gone.")
	heading("The tool belt")
	para("The Equipment window has a slot for each tool: pickaxe (mining), hatchet (woodcutting), sickle (farming), hammer (smelting and the anvil), knife (fletching) and needle (weaving and the tailor).")
	point("Better metal and higher quality work faster, up to 30% faster for a mithril tool at quality 100.")
	point("Better tools, and higher levels, gather and make better-quality materials.")
	point("Without the right tool, work is half speed, and what you gather is only quality 1 to 10.")
	point("Everyone starts with copper tools at quality 10. Stores sell more; the anvil makes them.")
}

func guideMaking() {
	para("Gear is made by hand. Visit an anvil, a fletcher or a tailor (or walk up to one) and use the Make tab. Pick a recipe, check the ingredients against your backpack, and make 1, 5 or as many as you can carry the materials for.")
	point("Bring the ingredients in your backpack: take them out at the bank first.")
	point("New gear comes out rough: around half your level, never better than its materials. Improve it to make it good.")
	point("A piece takes 10 seconds per bar at the anvil, or 8 per log, bar, leather or cloth at the fletcher and tailor, before skill and tool.")
	point("The tailor makes leather armour from leather (hunt deer, or take it from enemies) and cloth armour from cloth (weave cotton, or take it from enemies).")
	point("Without the station's tool on your belt, making is half speed and rougher.")
	recipeTables()
}

// recipeTables lists every gear recipe: a table per station, a row per
// item, grouped by metal or wood.
func recipeTables() {
	titles := map[resources.Building]string{
		resources.BuildingAnvil:    "Anvil (Smithing)",
		resources.BuildingFletcher: "Fletcher (Fletching)",
		resources.BuildingTailor:   "Tailor (Tailoring)",
	}
	for _, at := range []resources.Building{resources.BuildingAnvil, resources.BuildingFletcher, resources.BuildingTailor} {
		title := titles[at]
		heading(title)
		flags := imgui.TableFlagsRowBg | imgui.TableFlagsBordersInnerV | imgui.TableFlagsSizingFixedFit
		if !imgui.BeginTableV("recipes"+string(at), 3, flags, imgui.Vec2{}, 0) {
			continue
		}
		imgui.TableSetupColumn("Item")
		imgui.TableSetupColumn("Level")
		imgui.TableSetupColumn("Uses")
		imgui.TableHeadersRow()
		for _, r := range resources.GearRecipes {
			if r.At != at {
				continue
			}
			var uses []string
			for _, in := range r.Inputs {
				uses = append(uses, fmt.Sprintf("%d %s", in.Count, resources.ItemUnits[in.Item]))
			}
			imgui.TableNextRow()
			imgui.TableNextColumn()
			imgui.Text(resources.ItemNames[r.Item])
			imgui.TableNextColumn()
			imgui.Text(fmt.Sprint(r.Level))
			imgui.TableNextColumn()
			imgui.Text(strings.Join(uses, ", "))
		}
		imgui.EndTable()
	}
	heading("Also")
	point(fmt.Sprintf("Smelt (an action, at a blacksmith): 1 ore makes 1 bar. Fletch (an action, at a fletcher): 1 log makes %d arrows. Weave (an action, at a tailor): 2 cotton make 1 cloth.", resources.ArrowsPerLog))
}

func guideImproving() {
	para("Improving raises a piece's quality. Use the Improve tab at an anvil (metal gear and tools), a fletcher (bows and staves) or a tailor (leather and cloth armour), on a piece in your backpack.")
	point("You need the level to make that piece, its tool on your belt (hammer, knife or needle), and some of its main material (bars, logs, leather or cloth) of better quality than the piece.")
	point(fmt.Sprintf("Each attempt uses part of a unit of material: one bar or log lasts %d attempts. It picks the cheapest one that's still better than the piece.", resources.ImproveUnitAttempts))
	point(fmt.Sprintf("Each attempt wears your tool by %d%%. At 100%% it breaks.", resources.ImproveToolWear))
	point("Success raises the quality, never past the material's. A failure costs 1 quality.")
	point("The chance is 85% on pieces up to your level, falling 2% for each quality point above it, but those attempts give more XP.")
	point("Improving stops when you press Stop, walk away, run out of better material, break your tool, or get too tired.")
	heading("Tip")
	para("Better material lets a piece go higher. Gather with a good tool at a high level to get high-quality ore and logs, and smelt them yourself.")
}

func guideRepairing() {
	para("Gear and tools wear with use, shown as \"40% worn\": your weapon as you attack, your armour as you're hit, and tools as you improve with them. Better quality wears more slowly. Wear lowers how well a piece works, and at 100% it breaks and is gone; chat warns you at 75%. Repair it before then.")
	point("Repair works anywhere: use the Repair button next to a worn piece in the Equipment window, the Inventory window, or the Bank window at a bank.")
	point("A repair takes the wear back to 0%, but costs a little quality: about 1 for every 20% of wear.")
	point("Repair is a skill. Each repair trains it a little, and a higher level loses less quality, down to a third at level 99.")
	point("Hover over the Repair button to see what it will cost.")
}

func guideFighting() {
	para("Add a Fight action at a camp you've found. Your character rests at the bank until healthy, takes food, walks to the camp, and fights enemy after enemy.")
	point("Your weapon sets your style: swords use Melee and Strength; bows use Ranged and arrows from the ammo slot; staves use Magic. Bows and staves attack from up to 5 tiles away.")
	point("Necromancer staves heal you for a quarter of the damage they deal.")
	point("Armour depends on your class: warriors wear metal, archers leather, and mages and necromancers cloth. You can't wear another class's armour.")
	point("Your weapon wears as you attack and your armour as you're hit. Repair them before they break.")
	point("Defence makes enemies miss more often; Hitpoints raises your health.")
	point("XP is only paid when the enemy dies.")
	point("Enemies drop coins, and sometimes gear (rough, quality 5 to 25) or materials.")
	point("If you're defeated, you lose everything in your backpack and wake up back in the village. Your bank and coins are safe.")
	point("Camps further from the starting village are more dangerous.")
}

func guideHunting() {
	para("Deer wander the named forests. Add a Hunt action at a forest you've found and your character stalks the nearest deer, then tries to bring it down.")
	point("A catch gives 1 or 2 leather and a venison, and trains Hunting. Better Hunting means better leather and more catches.")
	point("A miss sends the deer running; your character picks another. Deer never fight back.")
	point("Venison is food: fighters take it from the bank like cabbages, and it heals 25.")
	point("Leather goes to a tailor, who makes it into leather armour for archers.")
}

func joinInts(ns []int) string {
	var parts []string
	for _, n := range ns {
		parts = append(parts, fmt.Sprint(n))
	}
	if len(parts) < 2 {
		return strings.Join(parts, "")
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " or " + parts[len(parts)-1]
}
