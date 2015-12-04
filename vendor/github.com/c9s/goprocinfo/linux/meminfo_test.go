package linux

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMemInfo(t *testing.T) {
	{
		var expected = MemInfo{1011048, 92096, 0, 44304, 681228, 4, 494100, 306804, 71424, 9576, 422676, 297228, 0, 0, 524284, 524280, 28, 0, 75444, 26384, 5624, 60884, 45068, 15816, 1112, 2936, 0, 0, 0, 1029808, 528152, 34359738367, 10504, 34359725792, 0, 0, 0, 0, 0, 0, 2048, 1056768, 0, 0}

		read, err := ReadMemInfo("proc/meminfo_1")
		if err != nil {
			t.Fatal("meminfo read fail")
		}
		t.Logf("%+v", read)

		if err := compareExpectedReadFieldsMemInfo(&expected, read); err != nil {
			t.Error(err.Error())
		}

		if !reflect.DeepEqual(*read, expected) {
			t.Error("not equal to expected")
		}
	}
	{
		var expected = MemInfo{132003228, 126199196, 130327756, 819908, 2910788, 0, 3043760, 1027084, 340788, 1056, 2702972, 1026028, 0, 0, 3903484, 3903484, 8, 0, 342276, 72380, 1704, 899472, 737432, 162040, 7328, 7120, 0, 0, 0, 69905096, 1024672, 34359738367, 495100, 34290957508, 0, 172032, 0, 0, 0, 0, 2048, 143652, 14501888, 121634816}

		read, err := ReadMemInfo("proc/meminfo_2")
		if err != nil {
			t.Fatal("meminfo read fail")
		}
		t.Logf("%+v", read)

		if err := compareExpectedReadFieldsMemInfo(&expected, read); err != nil {
			t.Error(err.Error())
		}

		if !reflect.DeepEqual(*read, expected) {
			t.Error("not equal to expected")
		}
	}
}

//This is a helper function which makes it easier to track down errors in expected versus read values.
func compareExpectedReadFieldsMemInfo(expected *MemInfo, read *MemInfo) error {
	elemExpected := reflect.ValueOf(*expected)
	typeOfElemExpected := elemExpected.Type()
	elemRead := reflect.ValueOf(*read)

	for i := 0; i < elemExpected.NumField(); i++ {
		fieldName := typeOfElemExpected.Field(i).Name

		if elemExpected.Field(i).Uint() != elemRead.Field(i).Uint() {
			return fmt.Errorf("Read value not equal to expected value for field %s. Got %d and expected %d.", fieldName, elemRead.Field(i).Uint(), elemExpected.Field(i).Uint())
		}
	}

	return nil
}
