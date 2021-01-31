package fabricate

import "reflect"

// A Registry tracks multiple Fabricators, so that one Fabricator can
// use a different Fabricator to generate a struct type field.
type Registry struct {
	Fabricators map[string]*Fabricator
	Config      *Config
}

// Define returns a new Fabricator registered to this Registry
func (r *Registry) Define(template interface{}) Fabricator {
	return Define(template).Register(r)
}

// NewRegistry returns a new Registry
func NewRegistry() *Registry {
	return &Registry{}
}

// Add registers an existing Fabricator to this Registry
func (r *Registry) Add(fabricator *Fabricator) {
	if r.Fabricators == nil {
		r.Fabricators = make(map[string]*Fabricator)
	}
	r.Fabricators[fabricator.name()] = fabricator
}

// SetConfig sets this Registry's Config
func (r *Registry) SetConfig(config *Config) *Registry {
	r.Config = config
	return r
}

func (f Fabricator) name() string {
	return reflect.ValueOf(f.Template).Type().Name()
}
