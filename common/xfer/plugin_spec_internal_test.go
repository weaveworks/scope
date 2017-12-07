package xfer

import (
	"fmt"
	"testing"

	"github.com/weaveworks/scope/test/reflect"
)

var benchmarkResult PluginSpecs

func TestMakePluginSpecs(t *testing.T) {
	for _, testcase := range []struct {
		inputs []string
		wants  []string
	}{
		{inputs: nil, wants: nil},
		{
			inputs: []string{"a"},
			wants:  []string{"a"},
		},
		{
			inputs: []string{"a", "a"},
			wants:  []string{"a"},
		},
		{
			inputs: []string{"b", "c", "a"},
			wants:  []string{"a", "b", "c"},
		},
	} {
		var inputs []PluginSpec
		for _, id := range testcase.inputs {
			inputs = append(inputs, PluginSpec{ID: id})
		}
		have := MakePluginSpecs(inputs...)
		var haveIDs []string
		have.ForEach(func(p PluginSpec) {
			haveIDs = append(haveIDs, p.ID)
		})
		if !reflect.DeepEqual(testcase.wants, haveIDs) {
			t.Errorf("%#v: want %#v, have %#v", inputs, testcase.wants, haveIDs)
		}
	}
}

func BenchmarkMakePluginSpecs(b *testing.B) {
	plugins := []PluginSpec{}
	for i := 1000; i >= 0; i-- {
		plugins = append(plugins, PluginSpec{ID: fmt.Sprint(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult = MakePluginSpecs(plugins...)
	}
}

func TestPluginSpecsAdd(t *testing.T) {
	for _, testcase := range []struct {
		input   PluginSpecs
		plugins []PluginSpec
		want    PluginSpecs
	}{
		{
			input:   PluginSpecs{},
			plugins: []PluginSpec{},
			want:    PluginSpecs{},
		},
		{
			input:   EmptyPluginSpecs,
			plugins: []PluginSpec{},
			want:    EmptyPluginSpecs,
		},
		{
			input:   MakePluginSpecs(PluginSpec{ID: "a"}),
			plugins: []PluginSpec{},
			want:    MakePluginSpecs(PluginSpec{ID: "a"}),
		},
		{
			input:   EmptyPluginSpecs,
			plugins: []PluginSpec{{ID: "a"}},
			want:    MakePluginSpecs(PluginSpec{ID: "a"}),
		},
		{
			input:   MakePluginSpecs(PluginSpec{ID: "a"}),
			plugins: []PluginSpec{{ID: "a"}},
			want:    MakePluginSpecs(PluginSpec{ID: "a"}),
		},
		{
			input: MakePluginSpecs(PluginSpec{ID: "b"}),
			plugins: []PluginSpec{
				{ID: "a"},
				{ID: "b"},
			},
			want: MakePluginSpecs(
				PluginSpec{ID: "a"},
				PluginSpec{ID: "b"},
			),
		},
		{
			input: MakePluginSpecs(PluginSpec{ID: "a"}),
			plugins: []PluginSpec{
				{ID: "c"},
				{ID: "b"},
			},
			want: MakePluginSpecs(
				PluginSpec{ID: "a"},
				PluginSpec{ID: "b"},
				PluginSpec{ID: "c"},
			),
		},
		{
			input: MakePluginSpecs(
				PluginSpec{ID: "a"},
				PluginSpec{ID: "c"},
			),
			plugins: []PluginSpec{
				{ID: "b"},
				{ID: "b"},
				{ID: "b"},
			},
			want: MakePluginSpecs(
				PluginSpec{ID: "a"},
				PluginSpec{ID: "b"},
				PluginSpec{ID: "c"},
			),
		},
	} {
		originalLen := testcase.input.Size()
		if want, have := testcase.want, testcase.input.Add(testcase.plugins...); !reflect.DeepEqual(want, have) {
			t.Errorf("%v + %v: want %v, have %v", testcase.input, testcase.plugins, want, have)
		}
		if testcase.input.Size() != originalLen {
			t.Errorf("%v + %v: modified the original input!", testcase.input, testcase.plugins)
		}
	}
}

func BenchmarkPluginSpecsAdd(b *testing.B) {
	n := EmptyPluginSpecs
	for i := 0; i < 600; i++ {
		n = n.Add(PluginSpec{ID: fmt.Sprint(i)})
	}

	plugin := PluginSpec{ID: "401.5"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult = n.Add(plugin)
	}
}

func TestPluginSpecsMerge(t *testing.T) {
	for _, testcase := range []struct {
		input PluginSpecs
		other PluginSpecs
		want  PluginSpecs
	}{
		{input: PluginSpecs{}, other: PluginSpecs{}, want: PluginSpecs{}},
		{input: EmptyPluginSpecs, other: EmptyPluginSpecs, want: EmptyPluginSpecs},
		{
			input: MakePluginSpecs(PluginSpec{ID: "a"}),
			other: EmptyPluginSpecs,
			want:  MakePluginSpecs(PluginSpec{ID: "a"}),
		},
		{
			input: EmptyPluginSpecs,
			other: MakePluginSpecs(PluginSpec{ID: "a"}),
			want:  MakePluginSpecs(PluginSpec{ID: "a"}),
		},
		{
			input: MakePluginSpecs(PluginSpec{ID: "a"}),
			other: MakePluginSpecs(PluginSpec{ID: "b"}),
			want:  MakePluginSpecs(PluginSpec{ID: "a"}, PluginSpec{ID: "b"}),
		},
		{
			input: MakePluginSpecs(PluginSpec{ID: "b"}),
			other: MakePluginSpecs(PluginSpec{ID: "a"}),
			want:  MakePluginSpecs(PluginSpec{ID: "a"}, PluginSpec{ID: "b"}),
		},
		{
			input: MakePluginSpecs(PluginSpec{ID: "a"}),
			other: MakePluginSpecs(PluginSpec{ID: "a"}),
			want:  MakePluginSpecs(PluginSpec{ID: "a"}),
		},
		{
			input: MakePluginSpecs(PluginSpec{ID: "a"}, PluginSpec{ID: "c"}),
			other: MakePluginSpecs(PluginSpec{ID: "a"}, PluginSpec{ID: "b"}),
			want:  MakePluginSpecs(PluginSpec{ID: "a"}, PluginSpec{ID: "b"}, PluginSpec{ID: "c"}),
		},
		{
			input: MakePluginSpecs(PluginSpec{ID: "b"}),
			other: MakePluginSpecs(PluginSpec{ID: "a"}),
			want:  MakePluginSpecs(PluginSpec{ID: "a"}, PluginSpec{ID: "b"}),
		},
	} {
		originalLen := testcase.input.Size()
		if want, have := testcase.want, testcase.input.Merge(testcase.other); !reflect.DeepEqual(want, have) {
			t.Errorf("%v + %v: want %v, have %v", testcase.input, testcase.other, want, have)
		}
		if testcase.input.Size() != originalLen {
			t.Errorf("%v + %v: modified the original input!", testcase.input, testcase.other)
		}
	}
}

func BenchmarkPluginSpecsMerge(b *testing.B) {
	n, other := PluginSpecs{}, PluginSpecs{}
	for i := 0; i < 600; i++ {
		n = n.Add(PluginSpec{ID: fmt.Sprint(i)})
	}

	for i := 400; i < 1000; i++ {
		other = other.Add(PluginSpec{ID: fmt.Sprint(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult = n.Merge(other)
	}
}
