package eve

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

// Decode parses the XML-encoded data and stores the result in the value pointed
// to by v, which must be an arbitrary struct.
//
// Because Decode uses the reflect package, it can only assign to exported
// (upper case) fields. Decode uses a case-sensitive comparison to match row
// sets to tag values and struct field names.
//
// Decode maps a row set to a struct by name or tag and passes the elements,
// which must be slices, to encoding/xml.Unmarshal. See encoding/xml.Unmarshal
// for more information on how the individual rows are parsed.
//
// An example of how to use these tags:
//    type ExampleRow struct {
//      ID          int    `xml:"groupID,attr"`
//      Name        string `xml:"name,attr"`
//      Description string `xml:"description,attr"`
//    }
//
//    type ExampleStruct struct {
//      CallGroups []*CallGroup `eve:"callGroups"`
//    }
func Decode(data []byte, v interface{}) (*Metadata, error) {
	return NewDecoder(bytes.NewReader(data)).Decode(v)
}

// NewDecoder creates a new EVE XML parser reading from r. If r does not
// implement io.ByteReader, encoding/xml will do its own buffering.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{XMLDecoder: xml.NewDecoder(r)}
}

// A Decoder represents an EVE XML parser reading a particular input stream.
type Decoder struct {
	// XMLDecoder is the underlying XML decoder for this Decoder instance.
	XMLDecoder *xml.Decoder

	// Fields are used by decodeRowSet to figure out which field a row set
	// belongs to.
	fields map[string]reflect.Value
}

// Decode reads tokens from the XML decoder to find metadata and row sets.
func (d *Decoder) Decode(v interface{}) (*Metadata, error) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return nil, errors.New("non-pointer passed to Decode")
	}

	// Load value from interface, but only if the result will be
	// usefully addressable.
	if val.Kind() == reflect.Interface && !val.IsNil() {
		e := val.Elem()
		if e.Kind() == reflect.Ptr && !e.IsNil() {
			val = e
		}
	}

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, errors.New("pointer to a non-struct passed to Decode")
	}

	d.fields = make(map[string]reflect.Value)
	t := val.Type()

	for i, j := 0, t.NumField(); i < j; i++ {
		field := t.Field(i)

		tag := field.Tag.Get("eve")
		if field.PkgPath != "" || tag == "-" {
			continue // Private field
		}

		if field.Type.Kind() != reflect.Slice {
			if tag == "" {
				// TODO: Word this better
				continue
			} else {
				return nil, errors.New("non-slice row acceptor passed to Decode")
			}
		}

		if tag == "" {
			tag = field.Name
		}

		d.fields[tag] = val.Field(i)
	}

	var expires string
	metadata := new(Metadata)
	for {
		token, err := d.XMLDecoder.Token()
		if err == io.EOF {
			break
		}

		t, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		startToken := &t

		switch startToken.Name.Local {
		case "cachedUntil":
			if err := d.XMLDecoder.DecodeElement(&expires, startToken); err != nil {
				return nil, err
			}
		case "rowset":
			setData, err := d.decodeRowSet(startToken)
			if err != nil {
				return nil, err
			}
			metadata.RowSets = append(metadata.RowSets, setData)
		}
	}

	cachedUntil, err := time.ParseInLocation(TimeFormat, expires, Timezone)
	if err != nil {
		return nil, err
	}
	metadata.Expires = cachedUntil

	return metadata, nil
}

func (d *Decoder) decodeRowSet(rowSetToken *xml.StartElement) (*RowSet, error) {
	setMetadata := new(RowSet)
	for _, attr := range rowSetToken.Attr {
		switch attr.Name.Local {
		case "name":
			setMetadata.Name = attr.Value
		case "key":
			setMetadata.Keys = strings.Split(attr.Value, ",")
		case "columns":
			setMetadata.Columns = strings.Split(attr.Value, ",")
		}
	}

	if setMetadata.Name == "" {
		return setMetadata, nil
	}

	field, ok := d.fields[setMetadata.Name]
	if !ok {
		return setMetadata, nil
	}
	typ := field.Type()

	for {
		token, err := d.XMLDecoder.Token()
		if err == io.EOF {
			return nil, errors.New("Expected end of rowset, got EOF instead")
		}

		if et, ok := token.(xml.EndElement); ok && et.Name.Local == "rowset" {
			return setMetadata, nil
		}

		st, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		startToken := &st

		if startToken.Name.Local == "row" {
			n := field.Len()
			if n >= field.Cap() {
				ncap := 2 * n
				if ncap < 4 {
					ncap = 4
				}
				new := reflect.MakeSlice(typ, n, ncap)
				reflect.Copy(new, field)
				field.Set(new)
			}
			field.SetLen(n + 1)

			if err := d.XMLDecoder.DecodeElement(field.Index(n).Addr().Interface(), startToken); err != nil {
				field.SetLen(n)
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("Unexpected token in row set: %s", startToken.Name.Local)
		}
	}
}
