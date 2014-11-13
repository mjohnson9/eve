package eve_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/nightexcessive/eve"
)

type TestSet struct {
	ID   int    `xml:"rowID,attr"`
	Name string `xml:"name,attr"`
}

type TestResult struct {
	Set1 []*TestSet `eve:"rowset1"`
	Set2 []*TestSet `eve:"rowset2"`
}

func TestDecoder(t *testing.T) {
	const testData = `
<?xml version='1.0' encoding='UTF-8'?>
<eveapi version="2">
  <currentTime>2006-01-02 15:04:05</currentTime>
  <result>
    <rowset name="rowset1" key="rowID" columns="rowID,name">
	  <row rowID="1" name="1-1"/>
	  <row rowID="2" name="1-2"/>
    </rowset>
    <rowset name="rowset2" key="rowID,name" columns="rowID,name">
	  <row rowID="1" name="2-1"/>
	  <row rowID="2" name="2-2"/>
    </rowset>
  </result>
  <cachedUntil>2020-01-02 15:04:05</cachedUntil>
</eveapi>`

	expectedMetadata := &eve.Metadata{
		Expires: time.Date(2020, time.January, 2, 15, 4, 5, 0, eve.Timezone),
		RowSets: []*eve.RowSet{
			&eve.RowSet{
				Name:    "rowset1",
				Keys:    []string{"rowID"},
				Columns: []string{"rowID", "name"},
			},
			&eve.RowSet{
				Name:    "rowset2",
				Keys:    []string{"rowID", "name"},
				Columns: []string{"rowID", "name"},
			},
		},
	}

	expectedOutput := &TestResult{
		Set1: []*TestSet{
			&TestSet{
				ID:   1,
				Name: "1-1",
			},
			&TestSet{
				ID:   2,
				Name: "1-2",
			},
		},
		Set2: []*TestSet{
			&TestSet{
				ID:   1,
				Name: "2-1",
			},
			&TestSet{
				ID:   2,
				Name: "2-2",
			},
		},
	}

	testOutput := new(TestResult)
	metadata, err := eve.Decode([]byte(testData), testOutput)
	if err != nil {
		t.Fatalf("Error in Decode: %s", err)
	}

	if !reflect.DeepEqual(metadata, expectedMetadata) {
		t.Errorf("metadata: got %+v, expected %+v", metadata, expectedMetadata)
	}

	if !reflect.DeepEqual(testOutput, expectedOutput) {
		t.Errorf("output: got %+v, expected %+v", testOutput, expectedOutput)
	}
}

func TestNonPointer(t *testing.T) {
	nonpointer := TestResult{}
	_, err := eve.Decode([]byte(""), nonpointer)
	if err == nil || err.Error() != "non-pointer passed to Decode" {
		t.Errorf("expected an error due to a non-pointer being passed to decode. got %s", err)
	}
}

func TestNonStruct(t *testing.T) {
	badValue := make([]string, 0, 5)
	_, err := eve.Decode([]byte(""), &badValue)
	if err == nil || err.Error() != "pointer to a non-struct passed to Decode" {
		t.Errorf("expected an error due to a non-struct being passed to decode. got %s", err)
	}
}
