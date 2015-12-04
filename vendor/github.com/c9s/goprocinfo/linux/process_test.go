package linux

import (
	"reflect"
	"testing"
)

func TestReadProcess(t *testing.T) {

	p, err := ReadProcess(3323, "proc")

	if err != nil {
		t.Fatal("process read fail", err)
	}

	expected := &Process{
		Status: ProcessStatus{
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
		},
		Statm: ProcessStatm{
			Size:     4053,
			Resident: 522,
			Share:    174,
			Text:     174,
			Lib:      0,
			Data:     286,
			Dirty:    0,
		},
		Stat: ProcessStat{
			Pid:                 3323,
			Comm:                "(proftpd)",
			State:               "S",
			Ppid:                1,
			Pgrp:                3323,
			Session:             3323,
			TtyNr:               0,
			Tpgid:               -1,
			Flags:               4202816,
			Minflt:              1311,
			Cminflt:             57367,
			Majflt:              0,
			Cmajflt:             1,
			Utime:               23,
			Stime:               58,
			Cutime:              24,
			Cstime:              49,
			Priority:            20,
			Nice:                0,
			NumThreads:          1,
			Itrealvalue:         0,
			Starttime:           2789,
			Vsize:               16601088,
			Rss:                 522,
			Rsslim:              4294967295,
			Startcode:           134512640,
			Endcode:             135222176,
			Startstack:          3217552592,
			Kstkesp:             3217551836,
			Kstkeip:             4118799382,
			Signal:              0,
			Blocked:             0,
			Sigignore:           272633856,
			Sigcatch:            8514799,
			Wchan:               0,
			Nswap:               0,
			Cnswap:              0,
			ExitSignal:          17,
			Processor:           7,
			RtPriority:          0,
			Policy:              0,
			DelayacctBlkioTicks: 1,
			GuestTime:           0,
			CguestTime:          0,
		},
		IO: ProcessIO{
			RChar:               3865585,
			WChar:               183294,
			Syscr:               6697,
			Syscw:               997,
			ReadBytes:           90112,
			WriteBytes:          45056,
			CancelledWriteBytes: 0,
		},
		Cmdline: "proftpd: (accepting connections)",
	}

	if !reflect.DeepEqual(p, expected) {
		t.Error("not equal to expected", expected)
	}

	t.Logf("%+v", p)
}
