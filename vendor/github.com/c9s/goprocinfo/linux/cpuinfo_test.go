package linux

import "testing"

func TestCPUInfo(t *testing.T) {

	cpuinfo, err := ReadCPUInfo("proc/cpuinfo")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", cpuinfo)

	if len(cpuinfo.Processors) != 8 {
		t.Fatal("wrong processor number : ", len(cpuinfo.Processors))
	}

	if cpuinfo.NumCore() != 8 {
		t.Fatal("wrong core number", cpuinfo.NumCore())
	}

	if cpuinfo.NumPhysicalCPU() != 2 {
		t.Fatal("wrong physical cpu number", cpuinfo.NumPhysicalCPU())
	}

	cpuinfo, err = ReadCPUInfo("proc/cpuinfo_2")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", cpuinfo)

	if len(cpuinfo.Processors) != 4 {
		t.Fatal("wrong processor number : ", len(cpuinfo.Processors))
	}

	if cpuinfo.NumCore() != 4 {
		t.Fatal("wrong core number", cpuinfo.NumCore())
	}

	// not sure at all here
	// does not match with https://github.com/randombit/cpuinfo/blob/master/x86/xeon_l5520
	if cpuinfo.NumPhysicalCPU() != 4 {
		t.Fatal("wrong physical cpu number", cpuinfo.NumPhysicalCPU())
	}

	cpuinfo, err = ReadCPUInfo("proc/cpuinfo_3")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", cpuinfo)

	if len(cpuinfo.Processors) != 4 {
		t.Fatal("wrong processor number : ", len(cpuinfo.Processors))
	}

	if cpuinfo.NumCore() != 2 {
		t.Fatal("wrong core number", cpuinfo.NumCore())
	}

	if cpuinfo.NumPhysicalCPU() != 1 {
		t.Fatal("wrong physical cpu number", cpuinfo.NumPhysicalCPU())
	}
}
