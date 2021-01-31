package fabricate

import (
	"reflect"
	"strings"
)

func (f Fabricator) setField(structField reflect.StructField, fieldValue reflect.Value, session Session) {
	generatedValue := f.generateValue(structField.Name, structField.Type, session)
	if !generatedValue.IsValid() {
		return
	}

	switch fieldValue.Kind().String() {
	case "int", "int8", "int16", "int32", "int64", "rune":
		fieldValue.SetInt(int64(generatedValue.Interface().(int)))
	case "uint", "uint8", "uint16", "uint32", "uint64", "byte", "uintptr":
		fieldValue.SetUint(uint64(generatedValue.Interface().(int)))
	case "float32", "float64":
		fieldValue.SetFloat(generatedValue.Interface().(float64))
	default:
		fieldValue.Set(generatedValue)
	}
}

func (f Fabricator) generateValue(fieldName string, fieldType reflect.Type, session Session) reflect.Value {
	// check if the field is exempt
	for _, omittedFieldName := range f.omittedFieldNames {
		if omittedFieldName == fieldName {
			return reflect.ValueOf(nil)
		}
	}

	// user-defined generator
	if generator := f.FieldGenerators[fieldName]; generator != nil {
		generatedField := generator(session)
		return reflect.ValueOf(generatedField)
	}

	// struct
	if fieldType.Kind().String() == "struct" {
		if f.Registry == nil {
			return reflect.ValueOf(nil)
		}

		if factory, ok := f.Registry.Fabricators[fieldType.Name()]; ok {
			fabricatorName := f.name()

			if _, ok := session.structGenerationCount[fabricatorName]; !ok {
				session.structGenerationCount[fabricatorName] = 0
			}
			session.structGenerationCount[fabricatorName]++
			if session.structGenerationCount[fabricatorName] > session.Config.MaxRecursiveStructDepth-1 {
				return reflect.ValueOf(nil)
			}

			return reflect.ValueOf(factory.fabricateWithSession(session, []string{})).Elem()
		}
	}

	// slice
	if fieldType.Kind().String() == "slice" {
		sliceType := fieldType.Elem()
		elemSlice := reflect.MakeSlice(reflect.SliceOf(sliceType), 0, session.Config.DefaultSliceLength)

		for i := 0; i < session.Config.DefaultSliceLength; i++ {
			generatedSliceElem := f.generateValue("", sliceType, session)
			if generatedSliceElem.IsValid() {
				elemSlice = reflect.Append(elemSlice, generatedSliceElem)
			}
		}

		return elemSlice
	}

	// ptr
	if fieldType.Kind().String() == "ptr" {
		if !session.Config.GeneratePtrs {
			return reflect.ValueOf(nil)
		}

		ptr := reflect.New(fieldType.Elem())

		generatedValue := f.generateValue("", fieldType.Elem(), session)
		if generatedValue.IsValid() {
			ptr.Elem().Set(generatedValue)
		}

		return ptr
	}

	// map
	if fieldType.Kind().String() == "map" {
		generatedMap := reflect.MakeMap(fieldType)
		for i := 0; i < session.Config.DefaultMapSize; i++ {
			generatedMapKey := f.generateValue("", fieldType.Key(), session)
			generatedMapValue := f.generateValue("", fieldType.Elem(), session)

			if generatedMapKey.IsValid() && generatedMapValue.IsValid() {
				generatedMap.SetMapIndex(generatedMapKey, generatedMapValue)
			}
		}

		return generatedMap
	}

	// finally, fall back on default generator (for composite types)
	if session.Config.DefaultGeneratorsEnabled {
		if defaultFieldGenerator := defaultFieldGenerator(fieldType.Name()); defaultFieldGenerator != nil {
			return reflect.ValueOf(defaultFieldGenerator(session))
		}
	}

	return reflect.ValueOf(nil)
}

func defaultFieldGenerator(fieldType string) FieldGenerator {
	switch fieldType {
	case "string":
		return func(session Session) interface{} {
			chars := []rune("abcdefghijklmnopqrstuvwxyz")
			var b strings.Builder
			for i := 0; i < session.Config.DefaultStringLength; i++ {
				b.WriteRune(chars[session.Rand.Intn(len(chars))])
			}

			return strings.Title(b.String())
		}
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "byte", "rune":
		return func(session Session) interface{} {
			return session.Rand.Intn(100) + 1
		}
	case "float32", "float64":
		return func(session Session) interface{} {
			return session.Rand.Float64() + 1
		}
	case "bool":
		return func(_ Session) interface{} {
			return true
		}
	}
	return nil
}
