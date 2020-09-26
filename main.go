package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
)

var (
	unsensed = flag.Bool("unsensed", false, "Generate unsensed maps")
	sensed   = flag.Bool("sensed", false, "Generate sensed maps")
)

type Map struct {
	angle, side, end []int // Involutions
}

func (m *Map) Ok() bool {
	// Check transitivity
	n := len(m.angle)
	v := make([]bool, n)
	v[0] = true
	cnt := 1
	for q := make([]int, 1, n); len(q) > 0; q = q[1:] {
		x := q[0]
		for _, y := range [3]int{m.angle[x], m.side[x], m.end[x]} {
			if !v[y] {
				v[y] = true
				cnt++
				q = append(q, y)
			}
		}
	}
	return cnt == n
}

func (m *Map) AddV(d int) {
	n := len(m.angle)

	for i := n; i < n+2*d; i++ {
		m.end = append(m.end, -1) // unknown yet
	}

	m.side = append(m.side, n+2*d-1)
	for i := n + 1; i < n+2*d-2; i += 2 {
		m.side = append(m.side, i+1)
		m.side = append(m.side, i)
	}
	m.side = append(m.side, n)

	for i := n; i < n+2*d; i += 2 {
		m.angle = append(m.angle, i+1)
		m.angle = append(m.angle, i)
	}
}

func (m Map) Order(x int) []int {
	n := len(m.angle)
	o := make([]int, n)
	for i := 0; i < n; i++ {
		o[i] = -1
	}
	q := make([]int, 1, n)
	q[0] = x
	o[x] = 0
	for cnt := 0; len(q) > 0; q = q[1:] {
		x := q[0]
		for _, y := range [3]int{m.angle[x], m.side[x], m.end[x]} {
			if o[y] == -1 {
				cnt++
				o[y] = cnt
				q = append(q, y)
			}
		}
	}
	return o
}

func (m Map) Rename(o []int) Map {
	n := len(m.angle)
	res := Map{
		angle: make([]int, n),
		side:  make([]int, n),
		end:   make([]int, n),
	}
	for i := 0; i < n; i++ {
		res.angle[o[i]] = o[m.angle[i]]
		res.side[o[i]] = o[m.side[i]]
		res.end[o[i]] = o[m.end[i]]
	}
	return res
}

func (m Map) Cmp(n Map) int {
	for i := 0; i < len(m.angle); i++ {
		if m.angle[i] != n.angle[i] {
			return m.angle[i] - n.angle[i]
		}
		if m.end[i] != n.end[i] {
			return m.end[i] - n.end[i]
		}
		if m.side[i] != n.side[i] {
			return m.side[i] - n.side[i]
		}
	}
	return 0
}

func (m Map) Less(n Map) bool {
	return m.Cmp(n) < 0
}

func (m Map) Eq(n Map) bool {
	return m.Cmp(n) == 0
}

func (m Map) Rooted() Map {
	return m.Rename(m.Order(0))
}

func (m Map) Unrooted(sensed bool) Map {
	var res Map
	side := make([]bool, 0)
	if sensed {
		side = m.Side()
	}
	for i := 0; i < len(m.angle); i++ {
		if sensed && !side[i] {
			continue
		}
		if n := m.Rename(m.Order(i)); i == 0 || n.Less(res) {
			res = n
		}
	}
	return res
}

func (m *Map) DelV(d int) {
	n := len(m.angle)
	m.end = m.end[:n-2*d]
	m.side = m.side[:n-2*d]
	m.angle = m.angle[:n-2*d]
}

func (m Map) Side() []bool {
	n := len(m.angle)
	v := make([]bool, n)
	v[0] = true
	for q := make([]int, 1, n); len(q) > 0; q = q[1:] {
		x := q[0]
		for _, y := range [2]int{m.side[m.angle[x]], m.side[m.end[x]]} {
			if !v[y] {
				v[y] = true
				q = append(q, y)
			}
		}
	}
	return v
}

func (m Map) Orientable() bool {
	return !m.Side()[1]
}

func (m Map) E() int {
	return len(m.angle) / 4
}

func (m Map) String() string {
	return fmt.Sprintf("a%v s%v e%v", m.angle, m.side, m.end, )
}

func cycles(p1, p2 []int) int {
	n := len(p1)
	v := make([]bool, n)
	cnt := 0
	for i := 0; i < n; i++ {
		for j := i; !v[j]; j = p2[p1[j]] {
			if j == i {
				cnt++
			}
			v[j] = true
			v[p1[j]] = true
		}
	}
	return cnt
}

func (m Map) V() int {
	return cycles(m.angle, m.side)
}

func (m Map) F() int {
	return cycles(m.angle, m.end)
}

func (m Map) Chi() int {
	return m.V() - m.E() + m.F()
}

func (m *Map) AddE(x, y int) {
	m.end[x] = y
	m.end[y] = x
	x, y = m.side[x], m.side[y]
	m.end[x] = y
	m.end[y] = x
}

func (m *Map) DelE(x, y int) {
	m.end[x] = -1
	m.end[y] = -1
	x, y = m.side[x], m.side[y]
	m.end[x] = -1
	m.end[y] = -1
}

func generateMaps(m *Map, degs *UIntMultiset, i int, out chan Map) {
	n := len(m.angle)
	if i == n && degs.Size() == 0 {
		// End generating
		if m.Ok() {
			out <- m.Rooted()
		}
		return
	}
	if i >= n {
		// Will result in unconnected map
		return
	}
	if m.end[i] != -1 {
		// Current flag already paired, continue
		generateMaps(m, degs, i+1, out)
		return
	}
	// Join to an existing vertex
	for j := i + 1; j < n; j++ {
		if m.end[j] == -1 && m.side[i] != j {
			m.AddE(i, j)
			generateMaps(m, degs, i+1, out)
			m.DelE(i, j)
		}
	}
	// Join to a new vertex
	for _, d := range degs.Distinct() {
		m.AddV(d)
		degs.Del(d)
		m.AddE(i, n)
		generateMaps(m, degs, i+1, out)
		m.DelE(i, n)
		degs.Add(d)
		m.DelV(d)
	}
}

func GenerateMaps(degs *UIntMultiset, out chan Map) {
	m := &Map{}
	for _, d := range degs.Distinct() {
		m.AddV(d)
		degs.Del(d)
		generateMaps(m, degs, 0, out)
		degs.Add(d)
		m.DelV(d)
	}
	close(out)
}

type UIntMultiset struct {
	vals            []int
	distinct, total int
}

func (im *UIntMultiset) Add(x int) {
	for len(im.vals) <= x {
		im.vals = append(im.vals, 0)
	}
	if im.vals[x] == 0 {
		im.distinct++
	}
	im.vals[x]++
	im.total++
}

func (im *UIntMultiset) Size() int {
	return im.total
}

func (im *UIntMultiset) Distinct() []int {
	res := make([]int, 0, im.distinct)
	for i, x := range im.vals {
		if x > 0 {
			res = append(res, i)
		}
	}
	return res
}

func (im *UIntMultiset) GetAndRemove() int {
	k := 0
	for im.vals[k] == 0 {
		k++
	}
	im.Del(k)
	return k
}

func (im *UIntMultiset) Del(k int) {
	if k > len(im.vals) || im.vals[k] <= 0 {
		panic(fmt.Sprintf("Nonexistent key: %d in %v", k, im.vals))
	}
	im.vals[k]--
	if im.vals[k] == 0 {
		im.distinct--
	}
	im.total--
}

func ParseDegreeMultiset(strs []string) (*UIntMultiset, error) {
	res := &UIntMultiset{}
	for _, s := range strs {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		res.Add(i)
	}
	return res, nil
}

func main() {
	flag.Parse()
	d, err := ParseDegreeMultiset(flag.Args())
	if err != nil {
		log.Fatalf("Can't parse degrees: %v", flag.Args())
	}
	if *sensed && *unsensed {
		log.Fatalf("Can't have both sensed and unsensed flags set")
	}
	unlabelled := *sensed || *unsensed
	c := make(chan Map)
	go GenerateMaps(d, c)
	for m := range c {
		if unlabelled && !m.Eq(m.Unrooted(*sensed)) {
			continue
		}
		if m.Orientable() {
			fmt.Printf("%d + %s\n", 1-m.Chi()/2, m.String())
		} else {
			fmt.Printf("%d - %s\n", 2-m.Chi(), m.String())
		}
	}
}
