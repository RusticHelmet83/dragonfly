package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

var (
	_ EntityInsider         = WoodPressurePlate{}
	_ world.ScheduledTicker = WoodPressurePlate{}
	_ EntityInsider         = StonePressurePlate{}
	_ world.ScheduledTicker = StonePressurePlate{}
	_ EntityInsider         = PolishedBlackstonePressurePlate{}
	_ world.ScheduledTicker = PolishedBlackstonePressurePlate{}
	_ EntityInsider         = LightWeightedPressurePlate{}
	_ world.ScheduledTicker = LightWeightedPressurePlate{}
	_ EntityInsider         = HeavyWeightedPressurePlate{}
	_ world.ScheduledTicker = HeavyWeightedPressurePlate{}
)

// WoodPressurePlate is a wooden pressure plate that provides redstone power and visual/audio feedback when stepped on.
type WoodPressurePlate struct {
	transparent
	flowingWaterDisplacer
	bass

	// Wood is the wood type of the pressure plate.
	Wood WoodType
	// Powered is whether the pressure plate is currently pressed down.
	Powered bool
}

// Model ...
func (p WoodPressurePlate) Model() world.BlockModel {
	return model.PressurePlate{Powered: p.Powered}
}

// SideClosed ...
func (WoodPressurePlate) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// FlammabilityInfo ...
func (p WoodPressurePlate) FlammabilityInfo() FlammabilityInfo {
	if !p.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 20, true)
}

// FuelInfo ...
func (p WoodPressurePlate) FuelInfo() item.FuelInfo {
	if !p.Wood.Flammable() {
		return item.FuelInfo{}
	}
	return newFuelInfo(time.Second * 15)
}

// BreakInfo ...
func (p WoodPressurePlate) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, axeEffective, oneOf(WoodPressurePlate{Wood: p.Wood}))
}

// RedstonePower ...
func (p WoodPressurePlate) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	if p.Powered {
		return 15
	}
	return 0
}

// RedstoneStrongPower ...
func (p WoodPressurePlate) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if p.Powered && face == cube.FaceDown {
		return 15
	}
	return 0
}

// EntityInside ...
func (p WoodPressurePlate) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if !p.Powered && pressurePlateTouching(pos, e) {
		p.Powered = true
		tx.SetBlock(pos, p, nil)
		tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOn{})
		tx.ScheduleBlockUpdate(pos, p, time.Millisecond*500)
	}
}

// ScheduledTick ...
func (p WoodPressurePlate) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !p.Powered {
		return
	}
	if pressurePlatePressed(pos, tx) {
		tx.ScheduleBlockUpdate(pos, p, time.Millisecond*500)
		return
	}
	p.Powered = false
	tx.SetBlock(pos, p, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOff{})
}

// NeighbourUpdateTick ...
func (p WoodPressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	supportPos := pos.Side(cube.FaceDown)
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, cube.FaceUp, tx) {
		breakBlock(p, pos, tx)
	}
}

// UseOnBlock ...
func (p WoodPressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, p)
	if !used || face != cube.FaceUp {
		return false
	}
	supportPos := pos.Side(cube.FaceDown)
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, cube.FaceUp, tx) {
		return false
	}
	p.Powered = false
	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (p WoodPressurePlate) EncodeItem() (name string, meta int16) {
	if p.Wood == OakWood() {
		return "minecraft:wooden_pressure_plate", 0
	}
	return "minecraft:" + p.Wood.String() + "_pressure_plate", 0
}

// EncodeBlock ...
func (p WoodPressurePlate) EncodeBlock() (name string, properties map[string]any) {
	signal := int32(0)
	if p.Powered {
		signal = 15
	}
	if p.Wood == OakWood() {
		return "minecraft:wooden_pressure_plate", map[string]any{"redstone_signal": signal}
	}
	return "minecraft:" + p.Wood.String() + "_pressure_plate", map[string]any{"redstone_signal": signal}
}

// allWoodPressurePlates returns a list of all wood pressure plate variants.
func allWoodPressurePlates() (plates []world.Block) {
	for _, w := range WoodTypes() {
		plates = append(plates, WoodPressurePlate{Wood: w, Powered: false})
		plates = append(plates, WoodPressurePlate{Wood: w, Powered: true})
	}
	return
}

// StonePressurePlate is a stone pressure plate that provides redstone power and visual/audio feedback when stepped on.
type StonePressurePlate struct {
	transparent
	flowingWaterDisplacer
	bassDrum

	// Powered is whether the pressure plate is currently pressed down.
	Powered bool
}

// Model ...
func (p StonePressurePlate) Model() world.BlockModel {
	return model.PressurePlate{Powered: p.Powered}
}

// SideClosed ...
func (StonePressurePlate) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// BreakInfo ...
func (p StonePressurePlate) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, pickaxeHarvestable, pickaxeEffective, oneOf(StonePressurePlate{}))
}

// RedstonePower ...
func (p StonePressurePlate) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	if p.Powered {
		return 15
	}
	return 0
}

// RedstoneStrongPower ...
func (p StonePressurePlate) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if p.Powered && face == cube.FaceDown {
		return 15
	}
	return 0
}

// EntityInside ...
func (p StonePressurePlate) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if !p.Powered && pressurePlateTouching(pos, e) {
		p.Powered = true
		tx.SetBlock(pos, p, nil)
		tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOn{})
		tx.ScheduleBlockUpdate(pos, p, time.Millisecond*500)
	}
}

// ScheduledTick ...
func (p StonePressurePlate) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !p.Powered {
		return
	}
	if pressurePlatePressed(pos, tx) {
		tx.ScheduleBlockUpdate(pos, p, time.Millisecond*500)
		return
	}
	p.Powered = false
	tx.SetBlock(pos, p, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOff{})
}

// NeighbourUpdateTick ...
func (p StonePressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	supportPos := pos.Side(cube.FaceDown)
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, cube.FaceUp, tx) {
		breakBlock(p, pos, tx)
	}
}

// UseOnBlock ...
func (p StonePressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, p)
	if !used || face != cube.FaceUp {
		return false
	}
	supportPos := pos.Side(cube.FaceDown)
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, cube.FaceUp, tx) {
		return false
	}
	p.Powered = false
	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (StonePressurePlate) EncodeItem() (name string, meta int16) {
	return "minecraft:stone_pressure_plate", 0
}

// EncodeBlock ...
func (p StonePressurePlate) EncodeBlock() (name string, properties map[string]any) {
	signal := int32(0)
	if p.Powered {
		signal = 15
	}
	return "minecraft:stone_pressure_plate", map[string]any{"redstone_signal": signal}
}

// allStonePressurePlates returns all stone pressure plate variants.
func allStonePressurePlates() []world.Block {
	return []world.Block{
		StonePressurePlate{Powered: false},
		StonePressurePlate{Powered: true},
	}
}

// PolishedBlackstonePressurePlate is a polished blackstone pressure plate.
type PolishedBlackstonePressurePlate struct {
	transparent
	flowingWaterDisplacer
	bassDrum

	// Powered is whether the pressure plate is currently pressed down.
	Powered bool
}

// Model ...
func (p PolishedBlackstonePressurePlate) Model() world.BlockModel {
	return model.PressurePlate{Powered: p.Powered}
}

// SideClosed ...
func (PolishedBlackstonePressurePlate) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// BreakInfo ...
func (p PolishedBlackstonePressurePlate) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, pickaxeHarvestable, pickaxeEffective, oneOf(PolishedBlackstonePressurePlate{}))
}

// RedstonePower ...
func (p PolishedBlackstonePressurePlate) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	if p.Powered {
		return 15
	}
	return 0
}

// RedstoneStrongPower ...
func (p PolishedBlackstonePressurePlate) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if p.Powered && face == cube.FaceDown {
		return 15
	}
	return 0
}

// EntityInside ...
func (p PolishedBlackstonePressurePlate) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if !p.Powered && pressurePlateTouching(pos, e) {
		p.Powered = true
		tx.SetBlock(pos, p, nil)
		tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOn{})
		tx.ScheduleBlockUpdate(pos, p, time.Millisecond*500)
	}
}

// ScheduledTick ...
func (p PolishedBlackstonePressurePlate) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !p.Powered {
		return
	}
	if pressurePlatePressed(pos, tx) {
		tx.ScheduleBlockUpdate(pos, p, time.Millisecond*500)
		return
	}
	p.Powered = false
	tx.SetBlock(pos, p, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOff{})
}

// NeighbourUpdateTick ...
func (p PolishedBlackstonePressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	supportPos := pos.Side(cube.FaceDown)
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, cube.FaceUp, tx) {
		breakBlock(p, pos, tx)
	}
}

// UseOnBlock ...
func (p PolishedBlackstonePressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, p)
	if !used || face != cube.FaceUp {
		return false
	}
	supportPos := pos.Side(cube.FaceDown)
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, cube.FaceUp, tx) {
		return false
	}
	p.Powered = false
	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (PolishedBlackstonePressurePlate) EncodeItem() (name string, meta int16) {
	return "minecraft:polished_blackstone_pressure_plate", 0
}

// EncodeBlock ...
func (p PolishedBlackstonePressurePlate) EncodeBlock() (name string, properties map[string]any) {
	signal := int32(0)
	if p.Powered {
		signal = 15
	}
	return "minecraft:polished_blackstone_pressure_plate", map[string]any{"redstone_signal": signal}
}

// allPolishedBlackstonePressurePlates returns all polished blackstone pressure plate variants.
func allPolishedBlackstonePressurePlates() []world.Block {
	return []world.Block{
		PolishedBlackstonePressurePlate{Powered: false},
		PolishedBlackstonePressurePlate{Powered: true},
	}
}

// LightWeightedPressurePlate is a gold pressure plate (light weighted) that activates when any entity steps on it.
type LightWeightedPressurePlate struct {
	transparent
	flowingWaterDisplacer
	bassDrum

	// Powered is whether the pressure plate is currently pressed down.
	Powered bool
}

// Model ...
func (p LightWeightedPressurePlate) Model() world.BlockModel {
	return model.PressurePlate{Powered: p.Powered}
}

// SideClosed ...
func (LightWeightedPressurePlate) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// BreakInfo ...
func (p LightWeightedPressurePlate) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, pickaxeHarvestable, pickaxeEffective, oneOf(LightWeightedPressurePlate{}))
}

// RedstonePower ...
func (p LightWeightedPressurePlate) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	if p.Powered {
		return 15
	}
	return 0
}

// RedstoneStrongPower ...
func (p LightWeightedPressurePlate) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if p.Powered && face == cube.FaceDown {
		return 15
	}
	return 0
}

// EntityInside ...
func (p LightWeightedPressurePlate) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if !p.Powered && pressurePlateTouching(pos, e) {
		p.Powered = true
		tx.SetBlock(pos, p, nil)
		tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOn{})
		tx.ScheduleBlockUpdate(pos, p, time.Millisecond*500)
	}
}

// ScheduledTick ...
func (p LightWeightedPressurePlate) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !p.Powered {
		return
	}
	if pressurePlatePressed(pos, tx) {
		tx.ScheduleBlockUpdate(pos, p, time.Millisecond*500)
		return
	}
	p.Powered = false
	tx.SetBlock(pos, p, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOff{})
}

// NeighbourUpdateTick ...
func (p LightWeightedPressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	supportPos := pos.Side(cube.FaceDown)
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, cube.FaceUp, tx) {
		breakBlock(p, pos, tx)
	}
}

// UseOnBlock ...
func (p LightWeightedPressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, p)
	if !used || face != cube.FaceUp {
		return false
	}
	supportPos := pos.Side(cube.FaceDown)
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, cube.FaceUp, tx) {
		return false
	}
	p.Powered = false
	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (LightWeightedPressurePlate) EncodeItem() (name string, meta int16) {
	return "minecraft:light_weighted_pressure_plate", 0
}

// EncodeBlock ...
func (p LightWeightedPressurePlate) EncodeBlock() (name string, properties map[string]any) {
	signal := int32(0)
	if p.Powered {
		signal = 15
	}
	return "minecraft:light_weighted_pressure_plate", map[string]any{"redstone_signal": signal}
}

// allLightWeightedPressurePlates returns all light weighted (gold) pressure plate variants.
func allLightWeightedPressurePlates() []world.Block {
	return []world.Block{
		LightWeightedPressurePlate{Powered: false},
		LightWeightedPressurePlate{Powered: true},
	}
}

// HeavyWeightedPressurePlate is an iron pressure plate (heavy weighted) that activates when any entity steps on it.
type HeavyWeightedPressurePlate struct {
	transparent
	flowingWaterDisplacer
	bassDrum

	// Powered is whether the pressure plate is currently pressed down.
	Powered bool
}

// Model ...
func (p HeavyWeightedPressurePlate) Model() world.BlockModel {
	return model.PressurePlate{Powered: p.Powered}
}

// SideClosed ...
func (HeavyWeightedPressurePlate) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// BreakInfo ...
func (p HeavyWeightedPressurePlate) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, pickaxeHarvestable, pickaxeEffective, oneOf(HeavyWeightedPressurePlate{}))
}

// RedstonePower ...
func (p HeavyWeightedPressurePlate) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	if p.Powered {
		return 15
	}
	return 0
}

// RedstoneStrongPower ...
func (p HeavyWeightedPressurePlate) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if p.Powered && face == cube.FaceDown {
		return 15
	}
	return 0
}

// EntityInside ...
func (p HeavyWeightedPressurePlate) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if !p.Powered && pressurePlateTouching(pos, e) {
		p.Powered = true
		tx.SetBlock(pos, p, nil)
		tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOn{})
		tx.ScheduleBlockUpdate(pos, p, time.Millisecond*500)
	}
}

// ScheduledTick ...
func (p HeavyWeightedPressurePlate) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !p.Powered {
		return
	}
	if pressurePlatePressed(pos, tx) {
		tx.ScheduleBlockUpdate(pos, p, time.Millisecond*500)
		return
	}
	p.Powered = false
	tx.SetBlock(pos, p, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOff{})
}

// NeighbourUpdateTick ...
func (p HeavyWeightedPressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	supportPos := pos.Side(cube.FaceDown)
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, cube.FaceUp, tx) {
		breakBlock(p, pos, tx)
	}
}

// UseOnBlock ...
func (p HeavyWeightedPressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, p)
	if !used || face != cube.FaceUp {
		return false
	}
	supportPos := pos.Side(cube.FaceDown)
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, cube.FaceUp, tx) {
		return false
	}
	p.Powered = false
	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (HeavyWeightedPressurePlate) EncodeItem() (name string, meta int16) {
	return "minecraft:heavy_weighted_pressure_plate", 0
}

// EncodeBlock ...
func (p HeavyWeightedPressurePlate) EncodeBlock() (name string, properties map[string]any) {
	signal := int32(0)
	if p.Powered {
		signal = 15
	}
	return "minecraft:heavy_weighted_pressure_plate", map[string]any{"redstone_signal": signal}
}

// allHeavyWeightedPressurePlates returns all heavy weighted (iron) pressure plate variants.
func allHeavyWeightedPressurePlates() []world.Block {
	return []world.Block{
		HeavyWeightedPressurePlate{Powered: false},
		HeavyWeightedPressurePlate{Powered: true},
	}
}

// pressurePlateDetectBox returns the area above the pressure plate at the position passed that an entity's
// bounding box must overlap to hold the plate down.
func pressurePlateDetectBox(pos cube.Pos) cube.BBox {
	return cube.Box(1.0/16.0, 0, 1.0/16.0, 15.0/16.0, 0.25, 15.0/16.0).Translate(pos.Vec3())
}

// pressurePlateTouching checks if the bounding box of the entity passed overlaps the detection area of the
// pressure plate at the position passed. Pressing and releasing a plate both use this check, so a plate can
// never be released while the entity that pressed it is still touching it.
func pressurePlateTouching(pos cube.Pos, e world.Entity) bool {
	return e.H().Type().BBox(e).Translate(e.Position()).IntersectsWith(pressurePlateDetectBox(pos))
}

// pressurePlatePressed checks if any entity is currently touching the pressure plate at the position passed.
func pressurePlatePressed(pos cube.Pos, tx *world.Tx) bool {
	// EntitiesWithin only yields entities whose position (their feet) lies inside the box, so the search area
	// is grown to also find entities whose bounding box overlaps the plate while their feet are outside of it.
	search := pressurePlateDetectBox(pos).GrowVec3(mgl64.Vec3{1, 3, 1})
	for e := range tx.EntitiesWithin(search) {
		if pressurePlateTouching(pos, e) {
			return true
		}
	}
	return false
}
