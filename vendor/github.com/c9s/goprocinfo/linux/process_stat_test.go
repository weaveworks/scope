package linux

import (
	"reflect"
	"testing"
)

func TestReadProcessStat(t *testing.T) {

	stat, err := ReadProcessStat("proc/3323/stat")

	if err != nil {
		t.Fatal("process stat read fail", err)
	}

	expected := &ProcessStat{
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
	}

	if !reflect.DeepEqual(stat, expected) {
		t.Errorf("not equal to expected %+v", expected)
	}

	t.Logf("%+v", stat)
}

func TestReadProcessStatWithSpace(t *testing.T) {

	stat, err := ReadProcessStat("proc/884/stat")

	if err != nil {
		t.Fatal("process stat read fail", err)
	}

	expected := &ProcessStat{
		Pid:                 884,
		Comm:                "(rs:main Q:Reg)",
		State:               "S",
		Ppid:                1,
		Pgrp:                873,
		Session:             873,
		TtyNr:               0,
		Tpgid:               -1,
		Flags:               4202816,
		Minflt:              561,
		Cminflt:             0,
		Majflt:              0,
		Cmajflt:             0,
		Utime:               68,
		Stime:               132,
		Cutime:              0,
		Cstime:              0,
		Priority:            20,
		Nice:                0,
		NumThreads:          4,
		Itrealvalue:         0,
		Starttime:           2161,
		Vsize:               255451136,
		Rss:                 409,
		Rsslim:              18446744073709551615,
		Startcode:           1,
		Endcode:             1,
		Startstack:          0,
		Kstkesp:             0,
		Kstkeip:             0,
		Signal:              0,
		Blocked:             2146172671,
		Sigignore:           16781830,
		Sigcatch:            1133601,
		Wchan:               18446744073709551615,
		Nswap:               0,
		Cnswap:              0,
		ExitSignal:          -1,
		Processor:           1,
		RtPriority:          0,
		Policy:              0,
		DelayacctBlkioTicks: 34,
		GuestTime:           0,
		CguestTime:          0,
	}

	if !reflect.DeepEqual(stat, expected) {
		t.Errorf("not equal to expected %+v", expected)
	}

	t.Logf("%+v", stat)
}

func TestReadProcessStatWithDoubleParentheses(t *testing.T) {

	stat, err := ReadProcessStat("proc/4854/stat")

	if err != nil {
		t.Fatal("process stat read fail", err)
	}

	expected := &ProcessStat{
		Pid:                 4854,
		Comm:                "((sd-pam))",
		State:               "S",
		Ppid:                4853,
		Pgrp:                4853,
		Session:             4853,
		TtyNr:               0,
		Tpgid:               -1,
		Flags:               1077944640,
		Minflt:              21,
		Cminflt:             0,
		Majflt:              0,
		Cmajflt:             0,
		Utime:               0,
		Stime:               0,
		Cutime:              0,
		Cstime:              0,
		Priority:            20,
		Nice:                0,
		NumThreads:          1,
		Itrealvalue:         0,
		Starttime:           4912,
		Vsize:               83931136,
		Rss:                 405,
		Rsslim:              18446744073709551615,
		Startcode:           1,
		Endcode:             1,
		Startstack:          0,
		Kstkesp:             0,
		Kstkeip:             0,
		Signal:              0,
		Blocked:             0,
		Sigignore:           4096,
		Sigcatch:            0,
		Wchan:               18446744073709551615,
		Nswap:               0,
		Cnswap:              0,
		ExitSignal:          17,
		Processor:           7,
		RtPriority:          0,
		Policy:              0,
		DelayacctBlkioTicks: 0,
		GuestTime:           0,
		CguestTime:          0,
	}

	if !reflect.DeepEqual(stat, expected) {
		t.Errorf("not equal to expected %+v", expected)
	}

	t.Logf("%+v", stat)

}
