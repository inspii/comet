package internal

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

func verifyCount(s *Sublist, count uint32, t *testing.T) {
	if s.Count() != count {
		t.Errorf("Count is %d, should be %d", s.Count(), count)
	}
}

func verifyLen(r []interface{}, l int, t *testing.T) {
	if len(r) != l {
		t.Errorf("Results len is %d, should be %d", len(r), l)
	}
}

func verifyMember(r []interface{}, val string, t *testing.T) {
	for _, v := range r {
		if v == nil {
			continue
		}
		if v.(string) == val {
			return
		}
	}
	t.Errorf("Value '%s' not found in results", val)
}

func verifyNumLevels(s *Sublist, expected int, t *testing.T) {
	dl := s.numLevels()
	if dl != expected {
		t.Errorf("NumLevels is %d, should be %d", dl, expected)
	}
}

func TestInit(t *testing.T) {
	s := NewSublist()
	verifyCount(s, 0, t)
}

func TestInsertCount(t *testing.T) {
	s := NewSublist()
	s.Insert("foo", "a")
	s.Insert("bar", "b")
	s.Insert("foo.bar", "b")
	verifyCount(s, 3, t)
}

func TestSimple(t *testing.T) {
	s := NewSublist()
	val := "a"
	sub := "foo"
	s.Insert(sub, val)
	r := s.Match(sub)
	verifyLen(r, 1, t)
	verifyMember(r, val, t)
}

func TestSimpleMultiTokens(t *testing.T) {
	s := NewSublist()
	val := "a"
	sub := "foo.bar.baz"
	s.Insert(sub, val)
	r := s.Match(sub)
	verifyLen(r, 1, t)
	verifyMember(r, val, t)
}

func TestPartialWildcard(t *testing.T) {
	s := NewSublist()
	literal := "a.b.c"
	pwc := "a.*.c"
	a, b := "a", "b"
	s.Insert(literal, a)
	s.Insert(pwc, b)
	r := s.Match(literal)
	verifyLen(r, 2, t)
	verifyMember(r, a, t)
	verifyMember(r, b, t)
}

func TestPartialWildcardAtEnd(t *testing.T) {
	s := NewSublist()
	literal := "a.b.c"
	pwc := "a.b.*"
	a, b := "a", "b"
	s.Insert(literal, a)
	s.Insert(pwc, b)
	r := s.Match(literal)
	verifyLen(r, 2, t)
	verifyMember(r, a, t)
	verifyMember(r, b, t)
}

func TestFullWildcard(t *testing.T) {
	s := NewSublist()
	literal := "a.b.c"
	fwc := "a.>"
	a, b := "a", "b"
	s.Insert(literal, a)
	s.Insert(fwc, b)
	r := s.Match(literal)
	verifyLen(r, 2, t)
	verifyMember(r, a, t)
	verifyMember(r, b, t)
}

func TestRemove(t *testing.T) {
	s := NewSublist()
	literal := "a.b.c.d"
	value := "foo"
	s.Insert(literal, value)
	verifyCount(s, 1, t)
	s.Remove(literal, "bar")
	verifyCount(s, 1, t)
	s.Remove("a.b.c", value)
	verifyCount(s, 1, t)
	s.Remove(literal, value)
	verifyCount(s, 0, t)
	r := s.Match(literal)
	verifyLen(r, 0, t)
}

func TestRemoveWildcard(t *testing.T) {
	s := NewSublist()
	literal := "a.b.c.d"
	pwc := "a.b.*.d"
	fwc := "a.b.>"
	value := "foo"
	s.Insert(pwc, value)
	s.Insert(fwc, value)
	s.Insert(literal, value)
	verifyCount(s, 3, t)
	r := s.Match(literal)
	verifyLen(r, 3, t)
	s.Remove(literal, value)
	verifyCount(s, 2, t)
	s.Remove(fwc, value)
	verifyCount(s, 1, t)
	s.Remove(pwc, value)
	verifyCount(s, 0, t)
}

func TestRemoveCleanup(t *testing.T) {
	s := NewSublist()
	literal := "a.b.c.d.e.f"
	depth := len(strings.Split(literal, "."))
	value := "foo"
	verifyNumLevels(s, 0, t)
	s.Insert(literal, value)
	verifyNumLevels(s, depth, t)
	s.Remove(literal, value)
	verifyNumLevels(s, 0, t)
}

func TestRemoveCleanupWildcards(t *testing.T) {
	s := NewSublist()
	literal := "a.b.*.d.e.>"
	depth := len(strings.Split(literal, "."))
	value := "foo"
	verifyNumLevels(s, 0, t)
	s.Insert(literal, value)
	verifyNumLevels(s, depth, t)
	s.Remove(literal, value)
	verifyNumLevels(s, 0, t)
}

func TestCacheBehavior(t *testing.T) {
	s := NewSublist()
	literal := "a.b.c"
	fwc := "a.>"
	a, b := "a", "b"
	s.Insert(literal, a)
	r := s.Match(literal)
	verifyLen(r, 1, t)
	s.Insert(fwc, b)
	r = s.Match(literal)
	verifyLen(r, 2, t)
	verifyMember(r, a, t)
	verifyMember(r, b, t)
	s.Remove(fwc, b)
	r = s.Match(literal)
	verifyLen(r, 1, t)
	verifyMember(r, a, t)
}

func checkBool(b, expected bool, t *testing.T) {
	if b != expected {
		debug.PrintStack()
		t.Fatalf("Expected %v, but got %v\n", expected, b)
	}
}

func TestMatchLiterals(t *testing.T) {
	checkBool(matchLiteral("foo", "foo"), true, t)
	checkBool(matchLiteral("foo", "bar"), false, t)
	checkBool(matchLiteral("foo", "*"), true, t)
	checkBool(matchLiteral("foo", ">"), true, t)
	checkBool(matchLiteral("foo.bar", ">"), true, t)
	checkBool(matchLiteral("foo.bar", "foo.>"), true, t)
	checkBool(matchLiteral("foo.bar", "bar.>"), false, t)
	checkBool(matchLiteral("stats.test.22", "stats.>"), true, t)
	checkBool(matchLiteral("stats.test.22", "stats.*.*"), true, t)
}

func TestCacheBounds(t *testing.T) {
	s := NewSublist()
	s.Insert("cache.>", "foo")

	tmpl := "cache.test.%d"
	loop := s.cmax + 100

	for i := 0; i < loop; i++ {
		sub := fmt.Sprintf(tmpl, i)
		s.Match(sub)
	}
	cs := int(len(s.cache))
	if cs > s.cmax {
		t.Fatalf("Cache is growing past limit: %d vs %d\n", cs, s.cmax)
	}
}

func TestStats(t *testing.T) {
	s := NewSublist()
	s.Insert("stats.>", "fwc")
	tmpl := "stats.test.%d"
	loop := 255
	total := uint32(loop + 1)

	for i := 0; i < loop; i++ {
		sub := fmt.Sprintf(tmpl, i)
		s.Insert(sub, "l")
	}

	stats := s.Stats()
	if time.Since(stats.StatsTime) > 50*time.Millisecond {
		t.Fatalf("StatsTime seems incorrect: %+v\n", stats.StatsTime)
	}
	if stats.NumSubs != total {
		t.Fatalf("Wrong stats for NumSubs: %d vs %d\n", stats.NumSubs, total)
	}
	if stats.NumInserts != uint64(total) {
		t.Fatalf("Wrong stats for NumInserts: %d vs %d\n", stats.NumInserts, total)
	}
	if stats.NumRemoves != 0 {
		t.Fatalf("Wrong stats for NumRemoves: %d vs %d\n", stats.NumRemoves, 0)
	}
	if stats.NumMatches != 0 {
		t.Fatalf("Wrong stats for NumMatches: %d vs %d\n", stats.NumMatches, 0)
	}

	for i := 0; i < loop; i++ {
		s.Match("stats.test.22")
	}
	s.Insert("stats.*.*", "pwc")
	s.Match("stats.test.22")

	stats = s.Stats()
	if stats.NumMatches != uint64(loop+1) {
		t.Fatalf("Wrong stats for NumMatches: %d vs %d\n", stats.NumMatches, loop+1)
	}
	expectedCacheHitRate := 255.0 / 256.0
	if stats.CacheHitRate != expectedCacheHitRate {
		t.Fatalf("Wrong stats for CacheHitRate: %.3g vs %0.3g\n", stats.CacheHitRate, expectedCacheHitRate)
	}
	if stats.MaxFanout != 3 {
		t.Fatalf("Wrong stats for MaxFanout: %d vs %d\n", stats.MaxFanout, 3)
	}
	if stats.AvgFanout != 2.5 {
		t.Fatalf("Wrong stats for MaxFanout: %f vs %f\n", stats.AvgFanout, 2.5)
	}

	s.ResetStats()
	stats = s.Stats()
	if time.Since(stats.StatsTime) > 50*time.Millisecond {
		t.Fatalf("After Reset: StatsTime seems incorrect: %+v\n", stats.StatsTime)
	}
	if stats.NumInserts != 0 {
		t.Fatalf("After Reset: Wrong stats for NumInserts: %d vs %d\n", stats.NumInserts, 0)
	}
	if stats.NumRemoves != 0 {
		t.Fatalf("After Reset: Wrong stats for NumRemoves: %d vs %d\n", stats.NumRemoves, 0)
	}
	if stats.NumMatches != 0 {
		t.Fatalf("After Reset: Wrong stats for NumMatches: %d vs %d\n", stats.NumMatches, 0)
	}
	if stats.CacheHitRate != 0.0 {
		t.Fatalf("After Reset: Wrong stats for CacheHitRate: %.3g vs %0.3g\n", stats.CacheHitRate, 0.0)
	}
}

// -- Benchmarks Setup --

var subs []string
var toks = []string{"apcera", "continuum", "component", "router", "api", "imgr", "jmgr", "auth"}
var sl = NewSublist()
var results = make([]interface{}, 0, 64)

func init() {
	subs = make([]string, 0, 256*1024)
	subsInit("")
	for i := 0; i < len(subs); i++ {
		sl.Insert(subs[i], subs[i])
	}
	addWildcards()
	println("Sublist holding ", sl.Count(), " subscriptions")
}

func subsInit(pre string) {
	var sub string
	for _, t := range toks {
		if len(pre) > 0 {
			sub = pre + "." + t
		} else {
			sub = t
		}
		subs = append(subs, sub)
		if len(strings.Split(sub, ".")) < 5 {
			subsInit(sub)
		}
	}
}

func addWildcards() {
	sl.Insert("cloud.>", "paas")
	sl.Insert("cloud.continuum.component.>", "health")
	sl.Insert("cloud.*.*.router.*", "traffic")
}

// -- Benchmarks Setup End --

func Benchmark______________________Insert(b *testing.B) {
	b.SetBytes(1)
	s := NewSublist()
	for i, l := 0, len(subs); i < b.N; i++ {
		index := i % l
		s.Insert(subs[index], subs[index])
	}
}

func Benchmark____________MatchSingleToken(b *testing.B) {
	b.SetBytes(1)
	s := "apcera"
	for i := 0; i < b.N; i++ {
		sl.Match(s)
	}
}

func Benchmark______________MatchTwoTokens(b *testing.B) {
	b.SetBytes(1)
	s := "apcera.continuum"
	for i := 0; i < b.N; i++ {
		sl.Match(s)
	}
}

func Benchmark_MatchFourTokensSingleResult(b *testing.B) {
	b.SetBytes(1)
	s := "apcera.continuum.component.router"
	for i := 0; i < b.N; i++ {
		sl.Match(s)
	}
}

func Benchmark_MatchFourTokensMultiResults(b *testing.B) {
	b.SetBytes(1)
	s := "cloud.continuum.component.router"
	for i := 0; i < b.N; i++ {
		sl.Match(s)
	}
}

func Benchmark_______MissOnLastTokenOfFive(b *testing.B) {
	b.SetBytes(1)
	s := "apcera.continuum.component.router.ZZZZ"
	for i := 0; i < b.N; i++ {
		sl.Match(s)
	}
}

func _BenchmarkRSS(b *testing.B) {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	println("HEAP:", m.HeapObjects)
	println("ALLOC:", m.Alloc)
	println("TOTAL ALLOC:", m.TotalAlloc)
	time.Sleep(30 * 1e9)
}
