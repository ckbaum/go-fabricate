package fabricate

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

// A Fabricator generates structs matching the type of its Template
type Fabricator struct {
	Template        interface{}
	FieldGenerators map[string]FieldGenerator
	Traits          map[string]Trait
	Registry        *Registry
	Config          *Config

	omittedFieldNames []string
	fieldNames        []string
}

// A FieldGenerator is a function that generates a particular field using the Session
type FieldGenerator func(Session) interface{}

// A Session stores data pertaining to a particular call to Fabricate, including the
// in-progress struct
type Session struct {
	Struct interface{}
	Config *Config
	Rand   *rand.Rand

	structGenerationCount map[string]int
}

// Define returns a new Fabricator for generating structs of the same type as the provided struct
func Define(template interface{}) Fabricator {
	typedValue := reflect.New(reflect.ValueOf(&template).Elem().Elem().Type()).Elem()
	fieldNames := []string{}

	for i := 0; i < typedValue.NumField(); i++ {
		fieldNames = append(fieldNames, typedValue.Type().Field(i).Name)
	}

	return Fabricator{
		Template:        template,
		FieldGenerators: map[string]FieldGenerator{},
		Traits:          map[string]Trait{},
		fieldNames:      fieldNames,
	}
}

// Fabricate generates a new struct as configured by this Fabricator. Its return value
// should be type asserted to a pointer.
func (f Fabricator) Fabricate(traitNames ...string) interface{} {
	session := Session{
		Rand:                  rand.New(rand.NewSource(time.Now().UnixNano())),
		structGenerationCount: map[string]int{},
	}

	// Set the Session's Config to the Config defined on the Fabricator, the Config defined
	// on the Fabricator's Registry, or the default	Config
	if f.Config != nil {
		session.Config = f.Config
	} else if f.Registry != nil && f.Registry.Config != nil {
		session.Config = f.Registry.Config
	} else {
		session.Config = defaultConfig()
	}

	return f.fabricateWithSession(session, traitNames)
}

// With returns a Fabricator that will use any non-zero values of the provided struct in the
// structs that it generates.
func (f Fabricator) With(newTemplate interface{}) Fabricator {
	f.Template = newTemplate
	return f
}

// WithConfig returns a Fabricator with the provided Config
func (f Fabricator) WithConfig(config *Config) Fabricator {
	f.Config = config
	return f
}

// Field returns a Fabricator that includes a function for generating a particular field.
func (f Fabricator) Field(fieldName string, generator FieldGenerator) Fabricator {
	f.FieldGenerators[fieldName] = generator

	// move this field to the end of the field generation order
	newFieldNames := []string{}
	for _, fn := range f.fieldNames {
		if fn != fieldName {
			newFieldNames = append(newFieldNames, fn)
		}
	}
	f.fieldNames = append(newFieldNames, fieldName)

	return f
}

// Fields returns a Fabricator that includes a function for generating multiple fields
func (f Fabricator) Fields(fieldNames []string, generator FieldGenerator) Fabricator {
	for _, fieldName := range fieldNames {
		f = f.Field(fieldName, generator)
	}
	return f
}

// ZeroFields returns a Fabricator that is set to generate a set of fields to their
// respective zero values
func (f Fabricator) ZeroFields(omitFieldNames ...string) Fabricator {
	newFieldNames := []string{}
	for _, fieldName := range f.fieldNames {
		omitted := false
		for _, omitFieldName := range omitFieldNames {
			if omitFieldName == fieldName {
				omitted = true
				break
			}
		}
		if !omitted {
			newFieldNames = append(newFieldNames, fieldName)
		}
	}
	f.fieldNames = newFieldNames

	return f
}

// Register returns a Fabricator that is associated with the provided Registry.
func (f Fabricator) Register(registry *Registry) Fabricator {
	f.Registry = registry
	registry.Add(&f)

	return f
}

// Trait defines a new Trait associated with this Fabricator
func (f Fabricator) Trait(traitName string) Trait {
	trait := Trait{FieldGenerators: make(map[string]FieldGenerator)}
	f.Traits[traitName] = trait
	return trait
}

func (f Fabricator) fabricateWithSession(session Session, traitNames []string) interface{} {
	for _, traitName := range traitNames {
		if trait, ok := f.Traits[traitName]; ok {
			for name, fieldGenerator := range trait.FieldGenerators {
				f.FieldGenerators[name] = fieldGenerator
			}
		} else {
			panic(fmt.Sprintf("Attempted to fabricate with an unregistered trait: %s", traitName))
		}
	}

	session.Struct = f.Template

	value := reflect.ValueOf(&session.Struct).Elem()
	typedValue := reflect.New(value.Elem().Type()).Elem()

	for _, fieldName := range f.fieldNames {
		fieldValue := typedValue.FieldByName(fieldName)
		fieldType, _ := typedValue.Type().FieldByName(fieldName)

		if fieldValue.IsZero() {
			f.setField(fieldType, fieldValue, session)
			value.Set(typedValue)
		}
	}

	value.Set(typedValue)
	ptr := reflect.New(reflect.TypeOf(session.Struct))
	ptr.Elem().Set(reflect.ValueOf(session.Struct))
	return ptr.Interface()
}
