package fabricate

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

type person struct {
	FirstName     string
	MiddleName    string
	LastName      string
	FullName      string
	Age           int
	IsSenior      bool
	Vehicle       vehicle
	FamilyMembers []*person
}

type vehicle struct {
	Year     int
	Make     string
	Color    string
	Wheels   int
	Nickname *string
}

type allTypes struct {
	Bool    bool
	String  string
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	UInt    uint
	UInt8   uint8
	UInt16  uint16
	UInt32  uint32
	UInt64  uint64
	UIntPtr uintptr
	Byte    byte
	Rune    rune
	Float32 float32
	Float64 float64

	Slice []int
	Ptr   *string
	Map   map[string]int
}

func TestTypeDefaults(t *testing.T) {
	allTypeFabricator := Define(allTypes{})

	fabricated := allTypeFabricator.Fabricate().(*allTypes)
	empty := allTypes{}

	// composites
	assert.NotEqual(t, fabricated.Bool, empty.Bool)
	assert.NotEqual(t, fabricated.String, empty.String)
	assert.NotEqual(t, fabricated.Int, empty.Int)
	assert.NotEqual(t, fabricated.Int8, empty.Int8)
	assert.NotEqual(t, fabricated.Int16, empty.Int16)
	assert.NotEqual(t, fabricated.Int32, empty.Int32)
	assert.NotEqual(t, fabricated.Int64, empty.Int64)
	assert.NotEqual(t, fabricated.UInt, empty.UInt)
	assert.NotEqual(t, fabricated.UInt8, empty.UInt8)
	assert.NotEqual(t, fabricated.UInt16, empty.UInt16)
	assert.NotEqual(t, fabricated.UInt32, empty.UInt32)
	assert.NotEqual(t, fabricated.UInt64, empty.UInt64)
	assert.NotEqual(t, fabricated.UIntPtr, empty.UIntPtr)
	assert.NotEqual(t, fabricated.Byte, empty.Byte)
	assert.NotEqual(t, fabricated.Rune, empty.Rune)
	assert.NotEqual(t, fabricated.Float32, empty.Float32)
	assert.NotEqual(t, fabricated.Float64, empty.Float64)

	// struct
	assert.Equal(t, 3, len(fabricated.Slice))
	assert.NotEqual(t, fabricated.Slice[0], fabricated.Slice[1])

	// ptr
	assert.NotEqual(t, *fabricated.Ptr, empty.String)

	// map
	assert.NotEqual(t, fabricated.Map, nil)
	assert.Equal(t, 3, len(fabricated.Map))
}

func TestFabricators(t *testing.T) {
	registry := NewRegistry()

	personFabricator := registry.Define(
		person{},
	).Field("IsSenior", func(_ Session) interface{} {
		return false
	}).Field("FullName", func(session Session) interface{} {
		return fmt.Sprintf(
			"%s %s",
			session.Struct.(person).FirstName,
			session.Struct.(person).LastName,
		)
	}).ZeroFields(
		"MiddleName",
	)

	vehicleFabricator := Define(
		vehicle{},
	).Field("Wheels", func(_ Session) interface{} {
		return 4
	}).Register(registry)

	personFabricator.Trait(
		"senior citizen",
	).Field("Age", func(_ Session) interface{} {
		return 70
	}).Field("IsSenior", func(_ Session) interface{} {
		return true
	})

	personFabricator.Trait(
		"motorcyclist",
	).Field("Age", func(_ Session) interface{} {
		return 21
	}).Field("vehicle", func(_ Session) interface{} {
		return vehicleFabricator.Fabricate("motorcycle")
	})

	vehicleFabricator.Trait(
		"motorcycle",
	).Field("Wheels", func(_ Session) interface{} {
		return 2
	})

	t.Run("Fields", func(t *testing.T) {
		p := personFabricator.With(person{FirstName: "John", LastName: "Doe"}).Fabricate().(*person)
		assert.Equal(t, fmt.Sprintf("%s %s", p.FirstName, p.LastName), p.FullName) // custom field generator
		assert.Equal(t, "", p.MiddleName)                                          // field omitted
		assert.Equal(t, 4, p.Vehicle.Wheels)                                       // from registered fabricator
	})

	t.Run("Traits", func(t *testing.T) {
		defaultPerson := personFabricator.Fabricate().(*person)
		senior := personFabricator.Fabricate("senior citizen").(*person)
		seniorMotorcyclist := personFabricator.Fabricate("motorcyclist", "senior citizen").(*person)

		assert.Equal(t, false, defaultPerson.IsSenior)
		assert.Equal(t, true, senior.IsSenior)
		assert.Equal(t, true, seniorMotorcyclist.IsSenior)

		assert.Equal(t, 70, senior.Age)
	})

	t.Run("Max recursive depth", func(t *testing.T) {
		p := personFabricator.Fabricate().(*person)
		assert.Equal(t, 3, len(p.FamilyMembers))
		assert.Equal(t, 0, len(p.FamilyMembers[0].FamilyMembers))
	})

	t.Run("Unregistered trait", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
			}
		}()

		_ = personFabricator.Fabricate("unregistered factory").(*person)

		t.Errorf("That was supposed to panic!")
	})
}

func TestConfigs(t *testing.T) {
	// test default config
	fabricator1 := Define(allTypes{})
	obj1 := fabricator1.Fabricate().(*allTypes)

	assert.NotEqual(t, 0, obj1.Int)
	assert.Equal(t, len(obj1.Map), 3)
	assert.Equal(t, len(obj1.Slice), 3)
	assert.Equal(t, len(obj1.String), 8)
	assert.NotEqual(t, obj1.Ptr, nil)

	// test custom config
	registry := NewRegistry()
	fabricator2 := fabricator1.WithConfig(&Config{
		DefaultGeneratorsEnabled: true,
		DefaultMapSize:           4,
		DefaultSliceLength:       4,
		DefaultStringLength:      9,
		GeneratePtrs:             false,
		MaxRecursiveStructDepth:  1,
	}).Register(registry)
	obj2 := fabricator2.Fabricate().(*allTypes)

	assert.Equal(t, 4, len(obj2.Map))
	assert.Equal(t, 4, len(obj2.Slice))
	assert.Equal(t, 9, len(obj2.String))
	assert.Equal(t, true, obj2.Ptr == nil)

	// test registry-set custom config
	registry.SetConfig(&Config{
		DefaultGeneratorsEnabled: true,
		DefaultMapSize:           4,
		DefaultSliceLength:       1,
		DefaultStringLength:      9,
		GeneratePtrs:             false,
		MaxRecursiveStructDepth:  1,
	})
	fabricator1 = fabricator1.Register(registry)
	obj3 := fabricator1.Fabricate().(*allTypes)
	assert.Equal(t, 1, len(obj3.Slice))
}

type user struct {
	FirstName string
	LastName  string
	Email     *string
	Verified  bool
	Friends   []*user
}

type author struct {
	Name    string
	Country string
	Books   []*book
}

type book struct {
	Title    string
	Language string
}

func TestDocExamples(t *testing.T) {
	userFabricator := Define(user{})

	u := userFabricator.Fabricate().(*user)
	assert.NotEqual(t, "", u.FirstName)
	assert.NotEqual(t, "", u.LastName)
	assert.NotEqual(t, 0, len(u.Friends))

	userFabricator = Define(
		user{},
	).Field("Verified", func(_ Session) interface{} {
		return true
	}).Field("FirstName", func(_ Session) interface{} {
		return "First"
	})

	u = userFabricator.Fabricate().(*user)
	assert.Equal(t, "First", u.FirstName)

	userFabricator = Define(
		user{},
	).Fields([]string{"FirstName", "LastName"}, func(_ Session) interface{} {
		return "Name"
	}).Field("Email", func(session Session) interface{} {
		email := fmt.Sprintf(
			"%s.%s@gmail.com",
			session.Struct.(user).FirstName,
			session.Struct.(user).LastName,
		)
		return &email
	})

	u = userFabricator.Fabricate().(*user)
	assert.Equal(t, "Name.Name@gmail.com", *u.Email)

	userFabricator.Trait(
		"no email",
	).Field("Email", func(_ Session) interface{} {
		return nil
	}).Field("Verified", func(_ Session) interface{} {
		return false
	})

	u = userFabricator.Fabricate("no email").(*user)
	assert.Nil(t, u.Email)
	assert.False(t, u.Verified)

	userFabricator.Trait(
		"short name",
	).Field("FirstName", func(_ Session) interface{} {
		return "Em"
	})

	u = userFabricator.Fabricate("no email", "short name").(*user)
	assert.Equal(t, "Em", u.FirstName)
	assert.Nil(t, u.Email)

	fabricatorRegistry := NewRegistry()

	authorFabricator := fabricatorRegistry.Define(
		author{},
	).Register(fabricatorRegistry)

	fabricatorRegistry.Define(
		book{},
	).Field("Language", func(_ Session) interface{} {
		return []string{"English", "French"}[rand.Intn(2)]
	}).Register(fabricatorRegistry)

	a := authorFabricator.Fabricate().(*author)
	assert.Equal(t, 3, len(a.Books))
	assert.True(t, a.Books[0].Language == "English" || a.Books[0].Language == "French")
}
