package main

import (
	"fmt"
	"strings"
	"math/rand"
)


/////////////////////////////////////////////////
// STRUCTS
type GameDelta float32
type GameEntityType int

const (
	GUI GameEntityType = iota
	PLAYER
	ENEMY
	PROJECTILE
)

type GameVariable struct {
	Type 		string
	StringValue string
	IntValue 	int
}

type GameSprite [][]rune
type GameEntityCoordinates struct {
	Y int
	X int
}
type GameEntityHitBox struct {
	Width int
	Height int
}
type GameEntityFields struct {
	sprite GameSprite
	x int
	y int
}
type GameEntity interface {
	Tick(*Game, GameDelta)
	Draw() (int, int, GameSprite)
	GetType() GameEntityType
	GetCoords() GameEntityCoordinates
	GetHitBox() GameEntityHitBox
}

type GameInputHandler interface{
	IsKeyPressed(string) bool
}


/////////////////////////////////////////////////
// FRAME BUFFER
type GameFrameBuffer struct{
	content [][]rune
	height int
	width int
}

func NewGameFrameBuffer(height int, width int) GameFrameBuffer {
	return GameFrameBuffer{
		height: height,
		width: width,
	}
}

func (gfb *GameFrameBuffer) Clear() {
	gfb.content = make([][]rune, gfb.height)
	for y := range gfb.content {
		gfb.content[y] = []rune(strings.Repeat(" ", gfb.width))
	}
}

func (gfb * GameFrameBuffer) addLine (y int, x int, line []rune) {
	for ix, char := range line {
		if !(ix+x <0 || ix+x >= gfb.width) {
			gfb.content[y][ix+x] = char
		}
	}
}

func (gfb *GameFrameBuffer) Add(y int, x int, content [][]rune) {
	for iy, line := range content {
		if !(iy+y < 0 || iy+y >= gfb.height) {
			gfb.addLine(iy+y, x, line)
		}
	}
}

func (gfb *GameFrameBuffer) Draw() {
	for _, line := range gfb.content {
		fmt.Println(string(line))
	}
}

func (gfb *GameFrameBuffer) ToString() string {
	s := ""

	for _, line := range gfb.content {
		s += string(line) + "\n"
	}

	return s
}


/////////////////////////////////////////////////
// GAME STRUCT
type Game struct {
	frameRate 		int
	entities 		[]GameEntity
	frameBuffer 	GameFrameBuffer
	inputHandler 	GameInputHandler
	globals  		map[string]GameVariable
}

func (g *Game) Tick(d GameDelta) {
	for _, entity := range g.entities {
		entity.Tick(g, d)
	}
}

func (g *Game) Draw() string {
	g.frameBuffer.Clear()

	for _, entity := range g.entities {
		y, x, sprite := entity.Draw()
		g.frameBuffer.Add(y, x, sprite)
	}

	return g.frameBuffer.ToString()
}

func (g *Game) Loop() {
	for true {
		g.Tick(1.0)
		g.Draw()
	}
}

func (g *Game) AddEntity(e GameEntity) {
	g.entities = append(g.entities, e)
}

func (g *Game) RemoveEntity(entityIndex int) {
	g.entities = append(g.entities[:entityIndex], g.entities[entityIndex+1:]...)
}

var game Game

func InitGame(inputHandler GameInputHandler) *Game {
	game = Game{
		frameRate: 60,
		entities: []GameEntity{},
		frameBuffer: NewGameFrameBuffer(40, 80),
		inputHandler: inputHandler,
		globals: map[string]GameVariable{
			"score": GameVariable{
				Type:"integer",
				IntValue: 0,
			},
			"lives_left": GameVariable {
				Type: "integer",
				IntValue: 3,
			},
		},
	}

	game.AddEntity(
		NewMyGameEntity(
			game.frameBuffer.height-7,
			game.frameBuffer.width/2,
		),
	)
	game.AddEntity(&ScoreBoardEntity{})	

	game.AddEntity(
		NewEnemyEntity(0,0),
	)
	game.AddEntity(
		NewEnemyEntity(0,7),
	)
	game.AddEntity(
		NewEnemyEntity(0,14),
	)
	for i := range [10]byte{} {
		game.AddEntity(
			NewEnemyEntity(0,i * 7),
		)
	}

	return &game
}

func GetGame() *Game {
	return &game
}


/////////////////////////////////////////////////
// Player
type PlayerEntity struct {
	coords GameEntityCoordinates
	sprite GameSprite
}

func (e *PlayerEntity) Move(dy int, dx int) {
	e.coords.Y += dy
	e.coords.X += dx
}

func (e *PlayerEntity) Tick(g *Game, d GameDelta) {
	if g.inputHandler.IsKeyPressed("right") {
		e.coords.X += 2
	}

	if g.inputHandler.IsKeyPressed("left") {
		e.coords.X += -2
	}

	if g.inputHandler.IsKeyPressed(" ") {
		projectile := NewProjectileEntity(
			e.coords.Y-1,
			e.coords.X+2,
		)
		g.AddEntity(projectile)
	}
}
func (e *PlayerEntity) Draw() (int, int, GameSprite) {
	return e.coords.Y, e.coords.X, e.sprite
}
func (e *PlayerEntity) GetType() GameEntityType {return PLAYER}
func (e *PlayerEntity) GetCoords() GameEntityCoordinates {return e.coords}
func (e *PlayerEntity) GetHitBox() GameEntityHitBox {
	return GameEntityHitBox{
		Height: len(e.sprite),
		Width: len(e.sprite[0]),
	}
}
 
func NewMyGameEntity(y int, x int) *PlayerEntity {
	return &PlayerEntity{
		coords: GameEntityCoordinates{
			Y: y,
			X: x,
		},
		sprite: [][]rune{
			[]rune("  ^  "),
			[]rune(" |o| "),
			[]rune("/ | \\"),
		},
	}
}

func IsNumBetween (num int, a int, b int) bool {
	return (num < a && num > b) || (num > a && num <b)
}

func IsCollide (a GameEntity, b GameEntity) bool {
	aCoords := a.GetCoords()
	aHitBox := a.GetHitBox()

	bCoords := b.GetCoords()
	bHitBox := b.GetHitBox()

	return (IsNumBetween(
		aCoords.X,
		bCoords.X,
		bCoords.X + bHitBox.Width,
	) || IsNumBetween(
		aCoords.X + aHitBox.Width,
		bCoords.X,
		bCoords.X + bHitBox.Width,
	)) && (IsNumBetween(
		aCoords.Y,
		bCoords.Y,
		bCoords.Y + bHitBox.Height,
	) || IsNumBetween(
		aCoords.Y + aHitBox.Height,
		bCoords.Y,
		bCoords.Y + bHitBox.Height,
	))
}

/////////////////////////////////////////////////
// Projectile
type ProjectileEntity struct {
	coords GameEntityCoordinates
	sprite GameSprite
}

func (e *ProjectileEntity) Tick(g *Game, d GameDelta) {
	e.coords.Y -= 1
	
	for index, entity := range g.entities {
		if entity.GetType() == ENEMY && IsCollide(e, entity) {
			g.RemoveEntity(index)
		
			scoreVar := g.globals["score"]
			scoreVar.IntValue += 1
			g.globals["score"] = scoreVar

			break
		}
	}
}
func (e *ProjectileEntity) Draw() (int, int, GameSprite) {
	return e.coords.Y, e.coords.X, e.sprite
}
func (e *ProjectileEntity) GetType() GameEntityType {return PROJECTILE}
func (e *ProjectileEntity) GetCoords() GameEntityCoordinates {return e.coords}
func (e *ProjectileEntity) GetHitBox() GameEntityHitBox {
	return GameEntityHitBox{
		Width:1,
		Height:1,
	}
}

func NewProjectileEntity(y int, x int) *ProjectileEntity {
	return &ProjectileEntity{
		coords: GameEntityCoordinates{
			Y: y,
			X: x,
		},
		sprite: [][]rune{
			[]rune("|"),
		},
	}
}

/////////////////////////////////////////////////
// Enemy
type EnemyEntity struct {
	coords GameEntityCoordinates
	sprite GameSprite
	velocity int
	counter int
}

func (e *EnemyEntity) Tick(g *Game, d GameDelta) {
	e.counter += 1
	if e.counter < 10 { return }

	if rand.Intn(20) == 1 {
		g.AddEntity(
			NewEnemyProjectileEntity(
				e.coords.Y + len(e.sprite) + 1,
				e.coords.X + len(e.sprite[0]) / 2,
			),
		)
	}

	e.coords.X += e.velocity
	e.counter = 0

	if e.coords.X >= g.frameBuffer.width - len(e.sprite[0]) {
		e.velocity = -1
		e.coords.Y += len(e.sprite) + 1
	}

	if e.coords.X <= 0 {
		e.velocity = 1
		e.coords.Y += len(e.sprite) + 1
	}	
}
func (e *EnemyEntity) Draw() (int, int, GameSprite) {
	return e.coords.Y, e.coords.X, e.sprite
}
func (e *EnemyEntity) GetType() GameEntityType {return ENEMY}
func (e *EnemyEntity) GetCoords() GameEntityCoordinates {return e.coords}
func (e *EnemyEntity) GetHitBox() GameEntityHitBox {
	return GameEntityHitBox{
		Height: len(e.sprite),
		Width: len(e.sprite[0]),
	}
}

func NewEnemyEntity(y int, x int) *EnemyEntity {
	return &EnemyEntity{
		coords: GameEntityCoordinates{
			Y: y,
			X: x,
		},
		sprite: [][]rune{
			[]rune(" ^ ^ "),
			[]rune("(000)"),
			[]rune("/! !\\"),
		},
		velocity: 1,
		counter: 0,
	}
}


type EnemyProjectileEntity struct {
	coords GameEntityCoordinates
	sprite GameSprite
}

func (e *EnemyProjectileEntity) Tick(g *Game, d GameDelta) {
	e.coords.Y += 1
	
	for index, entity := range g.entities {
		if entity.GetType() == PLAYER && IsCollide(e, entity) {
			// Remove Entity
			g.RemoveEntity(index)
		
			livesVar := g.globals["lives_left"]
			if livesVar.IntValue != 0 {
				livesVar.IntValue -= 1
				g.globals["lives_left"] = livesVar
				g.AddEntity(entity)
			}

			break
		}
	}
}
func (e *EnemyProjectileEntity) Draw() (int, int, GameSprite) {
	return e.coords.Y, e.coords.X, e.sprite
}
func (e *EnemyProjectileEntity) GetType() GameEntityType {return PROJECTILE}
func (e *EnemyProjectileEntity) GetCoords() GameEntityCoordinates {return e.coords}
func (e *EnemyProjectileEntity) GetHitBox() GameEntityHitBox {
	return GameEntityHitBox{
		Width:1,
		Height:1,
	}
}

func NewEnemyProjectileEntity(y int, x int) *EnemyProjectileEntity {
	return &EnemyProjectileEntity{
		coords: GameEntityCoordinates{
			Y: y,
			X: x,
		},
		sprite: [][]rune{
			[]rune("$"),
		},
	}
}



/////////////////////////////////////////////////
// Scoreboard
type ScoreBoardEntity struct {
	sprite [][]rune
	coords GameEntityCoordinates
}
func (e *ScoreBoardEntity) Tick(g *Game, d GameDelta) {
	e.coords.Y = game.frameBuffer.height - len(e.sprite)
}
func (e *ScoreBoardEntity) Draw() (int, int, GameSprite) {
	game := GetGame()
	score := game.globals["score"].IntValue
	livesLeft := game.globals["lives_left"].IntValue

	e.sprite = [][]rune{
		[]rune(strings.Repeat("#", game.frameBuffer.width)),
		[]rune(fmt.Sprintf("# Score: %v", score)),
		[]rune(fmt.Sprintf("# Lives Remaining: %v", livesLeft)),
		[]rune(strings.Repeat("#", game.frameBuffer.width)),
	}
	return e.coords.Y, e.coords.X, e.sprite
}

func (e *ScoreBoardEntity) GetType() GameEntityType {return GUI}
func (e *ScoreBoardEntity) GetCoords() GameEntityCoordinates {return e.coords}
func (e *ScoreBoardEntity) GetHitBox() GameEntityHitBox {
	return GameEntityHitBox{
		Height: 0,
		Width: 0,
	}
}