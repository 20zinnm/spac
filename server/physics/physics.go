package physics

import (
	"github.com/20zinnm/entity"
	"sync"
	"github.com/jakecoffman/cp"
	 "github.com/20zinnm/spac/common/world"
)

type Handler interface {
	Remove(entity.ID)
}

type HandlerFunc func(entity.ID)

func (fn HandlerFunc) Remove(entity entity.ID) {
	fn(entity)
}

type System struct {
	world      *world.World
	radius     float64
	handler    Handler
	entitiesMu sync.RWMutex
	entities   map[entity.ID]world.Component
}

func New(handler Handler, w *world.World, radius float64) *System {
	return &System{
		world:    w,
		radius:   radius,
		entities: make(map[entity.ID]world.Component),
		handler:  handler,
	}
}

func (s *System) Update(delta float64) {
	s.world.Lock()
	defer s.world.Unlock()

	s.world.Space.Step(delta)
	s.entitiesMu.RLock()
	for id, component := range s.entities {
		if !component.Position().Near(cp.Vector{}, s.radius) {
			s.handler.Remove(id)
		}
	}
	s.entitiesMu.RUnlock()
}

func (s *System) Add(entity entity.ID, component world.Component) {
	s.entitiesMu.Lock()
	s.entities[entity] = component
	s.entitiesMu.Unlock()
}

func (s *System) Remove(entity entity.ID) {
	s.entitiesMu.Lock()
	if c, ok := s.entities[entity]; ok {
		s.world.Lock()
		s.world.Space.RemoveBody(c.Body)
		s.world.Unlock()
		delete(s.entities, entity)
	}
	s.entitiesMu.Unlock()
}