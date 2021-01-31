package fabricate

// A Trait contains an alternate set of FieldGenerators that can be used
// for a particular Fabrication, in place of the Fabricator's FieldGenerators.
type Trait struct {
	FieldGenerators map[string]FieldGenerator
}

// Field returns a Trait that includes a function for generating a particular field.
func (t Trait) Field(fieldName string, generator FieldGenerator) Trait {
	t.FieldGenerators[fieldName] = generator

	return t
}
