package linux

import "testing"
import "reflect"
import "fmt"

func TestNetStat(t *testing.T) {
	{
		var expected = NetStat{0, 0, 1764, 180, 0, 0, 0, 0, 0, 0, 28321, 0, 0, 0, 0, 243, 25089, 53, 837, 0, 0, 95994, 623148353, 640988091, 0, 92391, 81263, 594305, 590571, 35, 6501, 81, 113, 213, 1, 223, 318, 1056, 287, 218, 6619, 435, 1, 975, 264, 17298, 871, 5836, 3843, 0, 0, 2, 520, 0, 0, 833, 0, 3235, 44, 0, 571, 163, 0, 138, 0, 0, 0, 19, 1312, 677, 129, 0, 0, 27986, 27713, 40522, 837, 0, 38648, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2772402103, 5189844022, 0, 0, 0, 0, 0, 0, 0, 0, 0}

		read, err := ReadNetStat("proc/net_netstat_1")
		if err != nil {
			t.Fatal("netstat read fail", err)
		}

		t.Logf("%+v", expected)
		t.Logf("%+v", read)

		if err := compareExpectedReadFieldsNetStat(&expected, read); err != nil {
			t.Error(err.Error())
		}

		if !reflect.DeepEqual(*read, expected) {
			t.Error("not equal to expected")
		}
	}
	{
		expected := NetStat{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 427717, 0, 0, 0, 0, 370, 3111446, 59, 1825, 0, 0, 170176, 0, 507248, 0, 47385919, 0, 7377770, 29264396, 0, 410, 0, 3, 1, 0, 0, 4, 0, 18, 579, 0, 2, 0, 50, 6, 949, 50, 332, 1005, 5893, 978, 0, 8, 0, 0, 1838, 8, 764, 0, 0, 157868, 3281, 0, 35, 0, 0, 0, 0, 28, 453, 46, 0, 0, 226, 316, 2725, 0, 0, 0, 0, 0, 0, 0, 0, 11292855, 88470, 0, 8, 261, 198, 0, 0, 0, 0, 0, 0, 0, 0, 842446, 118, 118, 11490, 859, 105365136, 0, 0, 0, 0, 249, 0, 205328912480, 353370957921, 0, 0, 92394, 0, 0, 157218430, 0, 0, 0}

		read, err := ReadNetStat("proc/net_netstat_2")
		if err != nil {
			t.Fatal("netstat read fail", err)
		}

		t.Logf("%+v", expected)
		t.Logf("%+v", read)

		if err := compareExpectedReadFieldsNetStat(&expected, read); err != nil {
			t.Error(err.Error())
		}

		if !reflect.DeepEqual(*read, expected) {
			t.Error("not equal to expected")
		}
	}
}

// This is a helper function which makes it easier to track down errors in expected versus read values.
func compareExpectedReadFieldsNetStat(expected *NetStat, read *NetStat) error {
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
