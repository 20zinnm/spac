package rendering

import (
	"github.com/faiface/pixel"
	"github.com/20zinnm/entity"
	"sync"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/imdraw"
	"math"
	"golang.org/x/image/colornames"
	"github.com/20zinnm/spac/common/world"
)

type System struct {
	win     *pixelgl.Window
	world   *world.World
	handler InputHandler
	// stateMu guards entities, camPos, tracking, canvas, and imd
	// It should be read-locked only for operations that do not modify any underlying state for any guarded variables, such as counting entities.
	stateMu  sync.RWMutex
	entities map[entity.ID]Renderable
	camPos   pixel.Vec
	tracking Trackable
	canvas   *pixelgl.Canvas
	imd      *imdraw.IMDraw
}

func New(win *pixelgl.Window, world *world.World, handler InputHandler) *System {
	return &System{
		win:      win,
		world:    world,
		handler:  handler,
		entities: make(map[entity.ID]Renderable),
		canvas:   pixelgl.NewCanvas(win.Bounds()),
		imd:      imdraw.New(nil),
	}
}

func (s *System) Add(entity entity.ID, renderable Renderable) {
	s.stateMu.Lock()
	s.entities[entity] = renderable
	s.stateMu.Unlock()
}

func (s *System) Update(delta float64) {
	s.stateMu.Lock()
	s.world.Lock()
	defer s.world.Unlock()
	defer s.stateMu.Unlock()
	targetPosn := pixel.Vec{}
	if s.tracking != nil {
		targetPosn = s.tracking.Position()
	}
	inputs := Inputs{
		Left:   s.win.Pressed(pixelgl.KeyA),
		Right:  s.win.Pressed(pixelgl.KeyD),
		Thrust: s.win.Pressed(pixelgl.KeyW),
		Shoot:  s.win.Pressed(pixelgl.MouseButton1) || s.win.Pressed(pixelgl.KeySpace),
	}
	s.handler.Handle(inputs)
	if s.win.Bounds() != s.canvas.Bounds() {
		s.canvas.SetBounds(s.win.Bounds())
	}
	s.camPos = pixel.Lerp(s.camPos, targetPosn, 1-math.Pow(1.0/128, delta))
	cam := pixel.IM.Moved(s.camPos.Scaled(-1))
	s.canvas.Clear(colornames.Black)
	s.imd.Clear()
	s.canvas.SetMatrix(cam)
	for _, entity := range s.entities {
		entity.Draw(s.imd)
	}
	s.imd.Draw(s.canvas)
	s.win.Clear(colornames.White)
	s.win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
		math.Max(
			s.win.Bounds().W()/s.canvas.Bounds().W(),
			s.win.Bounds().H()/s.canvas.Bounds().H(),
		),
	).Moved(s.win.Bounds().Center()))
	s.canvas.Draw(s.win, pixel.IM.Moved(s.canvas.Bounds().Center()))
}

func (s *System) Remove(entity entity.ID) {
	s.stateMu.Lock()
	delete(s.entities, entity)
	s.stateMu.Unlock()
}
