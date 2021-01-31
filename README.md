# go-fabricate

[![Go Report Card](https://goreportcard.com/badge/github.com/ckbaum/go-fabricate)](https://goreportcard.com/report/github.com/ckbaum/go-fabricate) [![PkgGoDev](https://pkg.go.dev/badge/github.com/ckbaum/go-fabricate)](https://pkg.go.dev/github.com/ckbaum/go-fabricate)

Expressive struct generation for your Go tests.

## Install

```bash
go get github.com/ckbaum/go-fabricate
```

## Usage and Examples

Let's define a `User` struct:

```go
type User struct {
  FirstName        string
  LastName         string
  Email            *string
  Verified         bool
  Friends          []*User
}
```

After defining a  `Fabricator`, fake `Users` can easily be generated with random, non-zero data.

```go
userFabricator := fabricate.Define(User{})

user := userFabricator.Fabricate().(*User)
// User{
//   FirstName: "Leqrewn",
//   LastName: "Rnzenfjq",
//   Email: &"Hnaeqwk",
//   Verified: true,
//   Friends: {&User{...}, &User{...},},
// }
```

With a barebones `Fabricator` like this one:

- All fields with composite data types will be set to a random value
- Pointers, slices and maps with composite element types will be populated
- Fields of the same type as the fabricated struct will also be generated using the fabricator (to a configurable depth.)

#### Custom Field behavior

After adding some custom field definitions, our `userFabricator` can generate more realistic users:

```go
userFabricator := fabricate.Define(
  User{},
).Field("Verified", func(_ fabricate.Session) interface{} {
  // have all our test users be verified
  return true
}).Field("FirstName", func(_ fabricate.Session) interface{} {
  // realistic names with help from a fake data package
  return gofakeit.FirstName()
})

user := userFabricator.Fabricate().(*User)
// User{
//   FirstName: "James",
//   LastName: "Poadarp",
//   Email: &"Ejasdklf",
//   Verified: true,
//   Friends: {&User{...}, &User{...},},
// }
```

Field definitions can reference the in-progress struct's other fields through the `Session` parameter. Here, `Email` references the in-progress struct (through `Session`) to incorporate the user's name.

```go
userFabricator := fabricate.Define(
  User{},
).Fields([]string{"FirstName", "LastName"}, func(_ fabricate.Session) interface{} {
  // realistic names with help from a fake data package
  // this time, FirstName and LastName share a generator
  return gofakeit.FirstName()
}).Field("Email", func(session fabricate.Session) interface{} {
  // an email based on this user's name
  email := fmt.Sprintf(
    "%s.%s@gmail.com",
    session.Struct.(*User).FirstName,
    session.Struct.(*User).LastName,
  )
  return &email
})
  
user := userFabricator.Fabricate().(*User)
// User{
//   FirstName: "James",
//   LastName: "Joyce",
//   Email: &"james.joyce@gmail.com",
//   Verified: true,
//   Friends: {&User{...}, &User{...},},
// }
```

#### Use Traits to describe groups of properties

Variations of Fabricators can be defined with Traits, and applied at fabrication time.

```go
userFabricator.Trait(
  "no email",
).Field("Email", func(_ fabricate.Session) interface{} {
  return nil
}).Field("Verified", func(_ fabricate.Session) interface{} {
  return false
})
  
user := userFabricator.Fabricate("no email").(*User)
// User{
//   FirstName: "James",
//   LastName: "Joyce",
//   Email: nil,
//   Verified: false,
//   Friends: {&User{...}, &User{...},},
// }
```

If you've defined multiple traits, they can be combined during fabrication:

```go
userFabricator.Trait(
  "short name",
).Field("FirstName", func(_ fabricate.Session) interface{} {
  return "Em"
})
  
user := userFabricator.Fabricate("no email", "short name").(*User)
// User{
//   FirstName: "Em",
//   LastName: "Joyce",
//   Email: nil,
//   Verified: false,
//   Friends: {&User{...}, &User{...},},
// }
```

#### Link fabricators together with a Registry

By defining a `Registry` that tracks multiple fabricators, nested Structs will automatically be generated using the appropriate `Fabricator`.

Consider two Structs , `Author` and `Book`

```go
type Author struct {
  Name string
  Country string
  Books []*Book
}

type Book struct {
  Title string
  Language string
}
```

This time, we'll define a `Registry` and define our two fabricators into the Registry.

```go
fabricatorRegistry := fabricate.NewRegistry()

authorFabricator := fabricatorRegistry.Define(
  Author{},
).Register(fabricatorRegistry)

bookFabricator := fabricatorRegistry.Define(
  Book{},
).Field("Language", func(_ fabricate.Session) interface{} {
  return []string{"English", "French"}[rand.Intn(2)]
}).Register(fabricatorRegistry)
```

Because these `Fabricators` are defined within the same `Registry`, the `Books` field in `Author` will be generated using `bookFabricator` by default.

```go
authorFabricator.Fabricate().(Author)
// Author{
//   Name: "Ejasdlf",
//   Country: "Pasfaf",
//   Books: []*Book{
//     &Book{
//       Title: "Peafee",
//       Language: "English"
//     },
//     &Book{
//       Title: "Peafee",
//       Language: "French"
//     },
//     &Book{
//       Title: "Peafee",
//       Language: "French"
//     },
// }
```

## License

The MIT License (MIT) - see LICENSE.md for more details
