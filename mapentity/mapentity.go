package mapentity

import (
	"encoding/json"
	"fmt"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"reflect"
	"strconv"
	"time"
)

type MapEntity map[string]interface{}

type NoIndexMapEntity map[string]interface{}

func (i MapEntity) Load(props []datastore.Property) error {
	return loadMap(i, props)
}

func (i MapEntity) Save() ([]datastore.Property, error) {
	var props []datastore.Property
	if err := saveMap(i, &props, false); err != nil {
		return nil, err
	}
	return props, nil
}

func (i NoIndexMapEntity) Load(props []datastore.Property) error {
	return loadMap(i, props)
}

func (i NoIndexMapEntity) Save() ([]datastore.Property, error) {
	var props []datastore.Property
	if err := saveMap(i, &props, true); err != nil {
		return nil, err
	}
	return props, nil
}

func loadMap(m map[string]interface{}, props []datastore.Property) error {
	for _, p := range props {
		if p.Multiple {
			var s reflect.Value
			if m[p.Name] == nil {
				t := reflect.SliceOf(reflect.TypeOf(p.Value))
				s = reflect.MakeSlice(t, 0, 0)
			} else {
				s = reflect.ValueOf(m[p.Name])
			}
			s = reflect.Append(s, reflect.ValueOf(p.Value))
			m[p.Name] = s.Interface()
		} else {
			m[p.Name] = p.Value
		}
	}
	return nil
}

func saveMap(m map[string]interface{}, props *[]datastore.Property, noIndex bool) error {
	for key, value := range m {
		v := reflect.ValueOf(value)
		if !v.IsValid() {
			continue
		}
		if v.Kind() == reflect.Slice && v.Type().Elem().Kind() != reflect.Uint8 {
			for j := 0; j < v.Len(); j++ {
				err := saveStructProperty(props, key, noIndex, true, v.Index(j))
				if err != nil {
					return err
				}
			}
			continue
		}
		// Otherwise, save the field itself.
		err := saveStructProperty(props, key, true, false, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func saveStructProperty(props *[]datastore.Property, name string, noIndex, multiple bool, v reflect.Value) error {
	p := datastore.Property{
		Name:     name,
		NoIndex:  noIndex,
		Multiple: multiple,
	}
	switch x := v.Interface().(type) {
	case json.Number:
		s := v.Interface().(json.Number).String()
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				p.Value = s
			} else {
				p.Value = f
			}
		} else {
			p.Value = i
		}
	case *datastore.Key:
		p.Value = x
	case time.Time:
		p.Value = x
	case appengine.BlobKey:
		p.Value = x
	case appengine.GeoPoint:
		p.Value = x
	case datastore.ByteString:
		p.Value = x
	default:
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			p.Value = v.Int()
		case reflect.Bool:
			p.Value = v.Bool()
		case reflect.String:
			p.Value = v.String()
		case reflect.Float32, reflect.Float64:
			p.Value = v.Float()
		case reflect.Slice:
			if v.Type().Elem().Kind() == reflect.Uint8 {
				p.NoIndex = true
				p.Value = v.Bytes()
			}
		case reflect.Struct:
			return fmt.Errorf("datastore: struct field is unsupported")
		}
	}
	if p.Value == nil {
		return fmt.Errorf("datastore: unsupported struct field type: %v", v.Type())
	}
	*props = append(*props, p)
	return nil
}
