package linux

import (
	"reflect"
	"testing"
)

func TestReadProcessStatus(t *testing.T) {

	status, err := ReadProcessStatus("proc/3323/status")

	if err != nil {
		t.Fatal("process io read fail", err)
	}

	expected := &ProcessStatus{
		Name:                     "proftpd",
		State:                    "S (sleeping)",
		Tgid:                     3323,
		Pid:                      3323,
		PPid:                     1,
		TracerPid:                0,
		RealUid:                  0,
		EffectiveUid:             111,
		SavedSetUid:              0,
		FilesystemUid:            111,
		RealGid:                  65534,
		EffectiveGid:             65534,
		SavedSetGid:              65534,
		FilesystemGid:            65534,
		FDSize:                   32,
		Groups:                   []int64{2001, 65534},
		VmPeak:                   16216,
		VmSize:                   16212,
		VmLck:                    0,
		VmHWM:                    2092,
		VmRSS:                    2088,
		VmData:                   872,
		VmStk:                    272,
		VmExe:                    696,
		VmLib:                    9416,
		VmPTE:                    36,
		VmSwap:                   0,
		Threads:                  1,
		SigQLength:               0,
		SigQLimit:                12091,
		SigPnd:                   0,
		ShdPnd:                   0,
		SigBlk:                   0,
		SigIgn:                   272633856,
		SigCgt:                   6450965743,
		CapInh:                   0,
		CapPrm:                   18446744073709551615,
		CapEff:                   0,
		CapBnd:                   18446744073709551615,
		Seccomp:                  0,
		CpusAllowed:              []uint32{255},
		VoluntaryCtxtSwitches:    5899,
		NonvoluntaryCtxtSwitches: 26,
	}

	if !reflect.DeepEqual(status, expected) {
		t.Error("not equal to expected", expected)
	}

	t.Logf("%+v", status)

}
