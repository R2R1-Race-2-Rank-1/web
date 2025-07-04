// Copyright 2011 Google Inc.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uuid

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

type test struct {
	in      string
	version Version
	variant Variant
	isuuid  bool
}

var tests = []test{
	{"f47ac10b-58cc-0372-8567-0e02b2c3d479", 0, RFC4122, true},
	{"f47ac10b-58cc-1372-8567-0e02b2c3d479", 1, RFC4122, true},
	{"f47ac10b-58cc-2372-8567-0e02b2c3d479", 2, RFC4122, true},
	{"f47ac10b-58cc-3372-8567-0e02b2c3d479", 3, RFC4122, true},
	{"f47ac10b-58cc-4372-8567-0e02b2c3d479", 4, RFC4122, true},
	{"f47ac10b-58cc-5372-8567-0e02b2c3d479", 5, RFC4122, true},
	{"f47ac10b-58cc-6372-8567-0e02b2c3d479", 6, RFC4122, true},
	{"f47ac10b-58cc-7372-8567-0e02b2c3d479", 7, RFC4122, true},
	{"f47ac10b-58cc-8372-8567-0e02b2c3d479", 8, RFC4122, true},
	{"f47ac10b-58cc-9372-8567-0e02b2c3d479", 9, RFC4122, true},
	{"f47ac10b-58cc-a372-8567-0e02b2c3d479", 10, RFC4122, true},
	{"f47ac10b-58cc-b372-8567-0e02b2c3d479", 11, RFC4122, true},
	{"f47ac10b-58cc-c372-8567-0e02b2c3d479", 12, RFC4122, true},
	{"f47ac10b-58cc-d372-8567-0e02b2c3d479", 13, RFC4122, true},
	{"f47ac10b-58cc-e372-8567-0e02b2c3d479", 14, RFC4122, true},
	{"f47ac10b-58cc-f372-8567-0e02b2c3d479", 15, RFC4122, true},

	{"urn:uuid:f47ac10b-58cc-4372-0567-0e02b2c3d479", 4, Reserved, true},
	{"URN:UUID:f47ac10b-58cc-4372-0567-0e02b2c3d479", 4, Reserved, true},
	{"f47ac10b-58cc-4372-0567-0e02b2c3d479", 4, Reserved, true},
	{"f47ac10b-58cc-4372-1567-0e02b2c3d479", 4, Reserved, true},
	{"f47ac10b-58cc-4372-2567-0e02b2c3d479", 4, Reserved, true},
	{"f47ac10b-58cc-4372-3567-0e02b2c3d479", 4, Reserved, true},
	{"f47ac10b-58cc-4372-4567-0e02b2c3d479", 4, Reserved, true},
	{"f47ac10b-58cc-4372-5567-0e02b2c3d479", 4, Reserved, true},
	{"f47ac10b-58cc-4372-6567-0e02b2c3d479", 4, Reserved, true},
	{"f47ac10b-58cc-4372-7567-0e02b2c3d479", 4, Reserved, true},
	{"f47ac10b-58cc-4372-8567-0e02b2c3d479", 4, RFC4122, true},
	{"f47ac10b-58cc-4372-9567-0e02b2c3d479", 4, RFC4122, true},
	{"f47ac10b-58cc-4372-a567-0e02b2c3d479", 4, RFC4122, true},
	{"f47ac10b-58cc-4372-b567-0e02b2c3d479", 4, RFC4122, true},
	{"f47ac10b-58cc-4372-c567-0e02b2c3d479", 4, Microsoft, true},
	{"f47ac10b-58cc-4372-d567-0e02b2c3d479", 4, Microsoft, true},
	{"f47ac10b-58cc-4372-e567-0e02b2c3d479", 4, Future, true},
	{"f47ac10b-58cc-4372-f567-0e02b2c3d479", 4, Future, true},

	{"f47ac10b158cc-5372-a567-0e02b2c3d479", 0, Invalid, false},
	{"f47ac10b-58cc25372-a567-0e02b2c3d479", 0, Invalid, false},
	{"f47ac10b-58cc-53723a567-0e02b2c3d479", 0, Invalid, false},
	{"f47ac10b-58cc-5372-a56740e02b2c3d479", 0, Invalid, false},
	{"f47ac10b-58cc-5372-a567-0e02-2c3d479", 0, Invalid, false},
	{"g47ac10b-58cc-4372-a567-0e02b2c3d479", 0, Invalid, false},
}

var constants = []struct {
	c    interface{}
	name string
}{
	{Person, "Person"},
	{Group, "Group"},
	{Org, "Org"},
	{Invalid, "Invalid"},
	{RFC4122, "RFC4122"},
	{Reserved, "Reserved"},
	{Microsoft, "Microsoft"},
	{Future, "Future"},
	{Domain(17), "Domain17"},
	{Variant(42), "BadVariant42"},
}

func testTest(t *testing.T, in string, tt test) {
	uuid := Parse(in)
	if ok := (uuid != nil); ok != tt.isuuid {
		t.Errorf("Parse(%s) got %v expected %v\b", in, ok, tt.isuuid)
	}
	if uuid == nil {
		return
	}

	if v := uuid.Variant(); v != tt.variant {
		t.Errorf("Variant(%s) got %d expected %d\b", in, v, tt.variant)
	}
	if v, _ := uuid.Version(); v != tt.version {
		t.Errorf("Version(%s) got %d expected %d\b", in, v, tt.version)
	}
}

func TestUUID(t *testing.T) {
	for _, tt := range tests {
		testTest(t, tt.in, tt)
		testTest(t, strings.ToUpper(tt.in), tt)
	}
}

func TestConstants(t *testing.T) {
	for x, tt := range constants {
		v, ok := tt.c.(fmt.Stringer)
		if !ok {
			t.Errorf("%x: %v: not a stringer", x, v)
		} else if s := v.String(); s != tt.name {
			v, _ := tt.c.(int)
			t.Errorf("%x: Constant %T:%d gives %q, expected %q", x, tt.c, v, s, tt.name)
		}
	}
}

func TestRandomUUID(t *testing.T) {
	m := make(map[string]bool)
	for x := 1; x < 32; x++ {
		uuid := NewRandom()
		s := uuid.String()
		if m[s] {
			t.Errorf("NewRandom returned duplicated UUID %s", s)
		}
		m[s] = true
		if v, _ := uuid.Version(); v != 4 {
			t.Errorf("Random UUID of version %s", v)
		}
		if uuid.Variant() != RFC4122 {
			t.Errorf("Random UUID is variant %d", uuid.Variant())
		}
	}
}

func TestNew(t *testing.T) {
	m := make(map[string]bool)
	for x := 1; x < 32; x++ {
		s := New()
		if m[s] {
			t.Errorf("New returned duplicated UUID %s", s)
		}
		m[s] = true
		uuid := Parse(s)
		if uuid == nil {
			t.Errorf("New returned %q which does not decode", s)
			continue
		}
		if v, _ := uuid.Version(); v != 4 {
			t.Errorf("Random UUID of version %s", v)
		}
		if uuid.Variant() != RFC4122 {
			t.Errorf("Random UUID is variant %d", uuid.Variant())
		}
	}
}

func clockSeq(t *testing.T, uuid UUID) int {
	seq, ok := uuid.ClockSequence()
	if !ok {
		t.Fatalf("%s: invalid clock sequence", uuid)
	}
	return seq
}

func TestClockSeq(t *testing.T) {
	// Fake time.Now for this test to return a monotonically advancing time; restore it at end.
	defer func(orig func() time.Time) { timeNow = orig }(timeNow)
	monTime := time.Now()
	timeNow = func() time.Time {
		monTime = monTime.Add(1 * time.Second)
		return monTime
	}

	SetClockSequence(-1)
	uuid1 := NewUUID()
	uuid2 := NewUUID()

	if clockSeq(t, uuid1) != clockSeq(t, uuid2) {
		t.Errorf("clock sequence %d != %d", clockSeq(t, uuid1), clockSeq(t, uuid2))
	}

	SetClockSequence(-1)
	uuid2 = NewUUID()

	// Just on the very off chance we generated the same sequence
	// two times we try again.
	if clockSeq(t, uuid1) == clockSeq(t, uuid2) {
		SetClockSequence(-1)
		uuid2 = NewUUID()
	}
	if clockSeq(t, uuid1) == clockSeq(t, uuid2) {
		t.Errorf("Duplicate clock sequence %d", clockSeq(t, uuid1))
	}

	SetClockSequence(0x1234)
	uuid1 = NewUUID()
	if seq := clockSeq(t, uuid1); seq != 0x1234 {
		t.Errorf("%s: expected seq 0x1234 got 0x%04x", uuid1, seq)
	}
}

func TestCoding(t *testing.T) {
	text := "7d444840-9dc0-11d1-b245-5ffdce74fad2"
	urn := "urn:uuid:7d444840-9dc0-11d1-b245-5ffdce74fad2"
	data := UUID{
		0x7d, 0x44, 0x48, 0x40,
		0x9d, 0xc0,
		0x11, 0xd1,
		0xb2, 0x45,
		0x5f, 0xfd, 0xce, 0x74, 0xfa, 0xd2,
	}
	if v := data.String(); v != text {
		t.Errorf("%x: encoded to %s, expected %s", data, v, text)
	}
	if v := data.URN(); v != urn {
		t.Errorf("%x: urn is %s, expected %s", data, v, urn)
	}

	uuid := Parse(text)
	if !Equal(uuid, data) {
		t.Errorf("%s: decoded to %s, expected %s", text, uuid, data)
	}
}

func TestVersion1(t *testing.T) {
	uuid1 := NewUUID()
	uuid2 := NewUUID()

	if Equal(uuid1, uuid2) {
		t.Errorf("%s:duplicate uuid", uuid1)
	}
	if v, _ := uuid1.Version(); v != 1 {
		t.Errorf("%s: version %s expected 1", uuid1, v)
	}
	if v, _ := uuid2.Version(); v != 1 {
		t.Errorf("%s: version %s expected 1", uuid2, v)
	}
	n1 := uuid1.NodeID()
	n2 := uuid2.NodeID()
	if !bytes.Equal(n1, n2) {
		t.Errorf("Different nodes %x != %x", n1, n2)
	}
	t1, ok := uuid1.Time()
	if !ok {
		t.Errorf("%s: invalid time", uuid1)
	}
	t2, ok := uuid2.Time()
	if !ok {
		t.Errorf("%s: invalid time", uuid2)
	}
	q1, ok := uuid1.ClockSequence()
	if !ok {
		t.Errorf("%s: invalid clock sequence", uuid1)
	}
	q2, ok := uuid2.ClockSequence()
	if !ok {
		t.Errorf("%s: invalid clock sequence", uuid2)
	}

	switch {
	case t1 == t2 && q1 == q2:
		t.Error("time stopped")
	case t1 > t2 && q1 == q2:
		t.Error("time reversed")
	case t1 < t2 && q1 != q2:
		t.Error("clock sequence chaned unexpectedly")
	}
}

func TestNode(t *testing.T) {
	// This test is mostly to make sure we don't leave nodeMu locked.
	ifname = ""
	if ni := NodeInterface(); ni != "" {
		t.Errorf("NodeInterface got %q, want %q", ni, "")
	}
	if SetNodeInterface("xyzzy") {
		t.Error("SetNodeInterface succeeded on a bad interface name")
	}
	if !SetNodeInterface("") {
		t.Error("SetNodeInterface failed")
	}
	if ni := NodeInterface(); ni == "" {
		t.Error("NodeInterface returned an empty string")
	}

	ni := NodeID()
	if len(ni) != 6 {
		t.Errorf("ni got %d bytes, want 6", len(ni))
	}
	hasData := false
	for _, b := range ni {
		if b != 0 {
			hasData = true
		}
	}
	if !hasData {
		t.Error("nodeid is all zeros")
	}

	id := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	SetNodeID(id)
	ni = NodeID()
	if !bytes.Equal(ni, id[:6]) {
		t.Errorf("got nodeid %v, want %v", ni, id[:6])
	}

	if ni := NodeInterface(); ni != "user" {
		t.Errorf("got inteface %q, want %q", ni, "user")
	}
}

func TestNodeAndTime(t *testing.T) {
	// Time is February 5, 1998 12:30:23.136364800 AM GMT

	uuid := Parse("7d444840-9dc0-11d1-b245-5ffdce74fad2")
	node := []byte{0x5f, 0xfd, 0xce, 0x74, 0xfa, 0xd2}

	ts, ok := uuid.Time()
	if ok {
		c := time.Unix(ts.UnixTime())
		want := time.Date(1998, 2, 5, 0, 30, 23, 136364800, time.UTC)
		if !c.Equal(want) {
			t.Errorf("Got time %v, want %v", c, want)
		}
	} else {
		t.Errorf("%s: bad time", uuid)
	}
	if !bytes.Equal(node, uuid.NodeID()) {
		t.Errorf("Expected node %v got %v", node, uuid.NodeID())
	}
}

func TestMD5(t *testing.T) {
	uuid := NewMD5(NameSpace_DNS, []byte("python.org")).String()
	want := "6fa459ea-ee8a-3ca4-894e-db77e160355e"
	if uuid != want {
		t.Errorf("MD5: got %q expected %q", uuid, want)
	}
}

func TestSHA1(t *testing.T) {
	uuid := NewSHA1(NameSpace_DNS, []byte("python.org")).String()
	want := "886313e1-3b8a-5372-9b90-0c9aee199e5d"
	if uuid != want {
		t.Errorf("SHA1: got %q expected %q", uuid, want)
	}
}

func TestNodeID(t *testing.T) {
	nid := []byte{1, 2, 3, 4, 5, 6}
	SetNodeInterface("")
	s := NodeInterface()
	if s == "" || s == "user" {
		t.Errorf("NodeInterface %q after SetInteface", s)
	}
	node1 := NodeID()
	if node1 == nil {
		t.Error("NodeID nil after SetNodeInterface", s)
	}
	SetNodeID(nid)
	s = NodeInterface()
	if s != "user" {
		t.Errorf("Expected NodeInterface %q got %q", "user", s)
	}
	node2 := NodeID()
	if node2 == nil {
		t.Error("NodeID nil after SetNodeID", s)
	}
	if bytes.Equal(node1, node2) {
		t.Error("NodeID not changed after SetNodeID", s)
	} else if !bytes.Equal(nid, node2) {
		t.Errorf("NodeID is %x, expected %x", node2, nid)
	}
}

func testDCE(t *testing.T, name string, uuid UUID, domain Domain, id uint32) {
	if uuid == nil {
		t.Errorf("%s failed", name)
		return
	}
	if v, _ := uuid.Version(); v != 2 {
		t.Errorf("%s: %s: expected version 2, got %s", name, uuid, v)
		return
	}
	if v, ok := uuid.Domain(); !ok || v != domain {
		if !ok {
			t.Errorf("%s: %d: Domain failed", name, uuid)
		} else {
			t.Errorf("%s: %s: expected domain %d, got %d", name, uuid, domain, v)
		}
	}
	if v, ok := uuid.Id(); !ok || v != id {
		if !ok {
			t.Errorf("%s: %d: Id failed", name, uuid)
		} else {
			t.Errorf("%s: %s: expected id %d, got %d", name, uuid, id, v)
		}
	}
}

func TestDCE(t *testing.T) {
	testDCE(t, "NewDCESecurity", NewDCESecurity(42, 12345678), 42, 12345678)
	testDCE(t, "NewDCEPerson", NewDCEPerson(), Person, uint32(os.Getuid()))
	testDCE(t, "NewDCEGroup", NewDCEGroup(), Group, uint32(os.Getgid()))
}

type badRand struct{}

func (r badRand) Read(buf []byte) (int, error) {
	for i, _ := range buf {
		buf[i] = byte(i)
	}
	return len(buf), nil
}

func TestBadRand(t *testing.T) {
	SetRand(badRand{})
	uuid1 := New()
	uuid2 := New()
	if uuid1 != uuid2 {
		t.Errorf("expected duplicates, got %q and %q", uuid1, uuid2)
	}
	SetRand(nil)
	uuid1 = New()
	uuid2 = New()
	if uuid1 == uuid2 {
		t.Errorf("unexpected duplicates, got %q", uuid1)
	}
}

func TestUUID_Array(t *testing.T) {
	expect := Array{
		0xf4, 0x7a, 0xc1, 0x0b,
		0x58, 0xcc,
		0x03, 0x72,
		0x85, 0x67,
		0x0e, 0x02, 0xb2, 0xc3, 0xd4, 0x79,
	}
	uuid := Parse("f47ac10b-58cc-0372-8567-0e02b2c3d479")
	if uuid == nil {
		t.Fatal("invalid uuid")
	}
	if uuid.Array() != expect {
		t.Fatal("invalid array")
	}
}

func TestArray_UUID(t *testing.T) {
	array := Array{
		0xf4, 0x7a, 0xc1, 0x0b,
		0x58, 0xcc,
		0x03, 0x72,
		0x85, 0x67,
		0x0e, 0x02, 0xb2, 0xc3, 0xd4, 0x79,
	}
	expect := Parse("f47ac10b-58cc-0372-8567-0e02b2c3d479")
	if expect == nil {
		t.Fatal("invalid uuid")
	}
	if !bytes.Equal(array.UUID(), expect) {
		t.Fatal("invalid uuid")
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		uuid := Parse("f47ac10b-58cc-0372-8567-0e02b2c3d479")
		if uuid == nil {
			b.Fatal("invalid uuid")
		}
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

func BenchmarkUUID_String(b *testing.B) {
	uuid := Parse("f47ac10b-58cc-0372-8567-0e02b2c3d479")
	if uuid == nil {
		b.Fatal("invalid uuid")
	}
	for i := 0; i < b.N; i++ {
		if uuid.String() == "" {
			b.Fatal("invalid uuid")
		}
	}
}

func BenchmarkUUID_URN(b *testing.B) {
	uuid := Parse("f47ac10b-58cc-0372-8567-0e02b2c3d479")
	if uuid == nil {
		b.Fatal("invalid uuid")
	}
	for i := 0; i < b.N; i++ {
		if uuid.URN() == "" {
			b.Fatal("invalid uuid")
		}
	}
}

func BenchmarkUUID_Array(b *testing.B) {
	expect := Array{
		0xf4, 0x7a, 0xc1, 0x0b,
		0x58, 0xcc,
		0x03, 0x72,
		0x85, 0x67,
		0x0e, 0x02, 0xb2, 0xc3, 0xd4, 0x79,
	}
	uuid := Parse("f47ac10b-58cc-0372-8567-0e02b2c3d479")
	if uuid == nil {
		b.Fatal("invalid uuid")
	}
	for i := 0; i < b.N; i++ {
		if uuid.Array() != expect {
			b.Fatal("invalid array")
		}
	}
}

func BenchmarkArray_UUID(b *testing.B) {
	array := Array{
		0xf4, 0x7a, 0xc1, 0x0b,
		0x58, 0xcc,
		0x03, 0x72,
		0x85, 0x67,
		0x0e, 0x02, 0xb2, 0xc3, 0xd4, 0x79,
	}
	expect := Parse("f47ac10b-58cc-0372-8567-0e02b2c3d479")
	if expect == nil {
		b.Fatal("invalid uuid")
	}
	for i := 0; i < b.N; i++ {
		if !bytes.Equal(array.UUID(), expect) {
			b.Fatal("invalid uuid")
		}
	}
}
