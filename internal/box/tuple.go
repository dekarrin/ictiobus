package box

// Pair is a simple struct that holds two values.
type Pair[E1, E2 any] struct {
	First  E1
	Second E2
}

func PairOf[E1, E2 any](first E1, second E2) Pair[E1, E2] {
	return Pair[E1, E2]{First: first, Second: second}
}

func (p Pair[E1, E2]) Elements() []interface{} {
	return []interface{}{p.First, p.Second}
}

func (p Pair[E1, E2]) All() (E1, E2) {
	return p.First, p.Second
}

// HPair is a homogenous struct that holds two values of the same type.
type HPair[E any] struct {
	First  E
	Second E
}

func HPairOf[E any](first E, second E) HPair[E] {
	return HPair[E]{First: first, Second: second}
}

func (hp HPair[E]) Elements() []E {
	return []E{hp.First, hp.Second}
}

func (hp HPair[E]) All() (E, E) {
	return hp.First, hp.Second
}

// Triple is a simple struct that holds three values.
type Triple[E1, E2, E3 any] struct {
	First  E1
	Second E2
	Third  E3
}

func NewTriple[E1, E2, E3 any](first E1, second E2, third E3) Triple[E1, E2, E3] {
	return Triple[E1, E2, E3]{First: first, Second: second, Third: third}
}

func (t Triple[E1, E2, E3]) Elements() []interface{} {
	return []interface{}{t.First, t.Second, t.Third}
}

func (t Triple[E1, E2, E3]) All() (E1, E2, E3) {
	return t.First, t.Second, t.Third
}

// HTriple is a homogenous struct that holds three values of the same type.
type HTriple[E any] struct {
	First  E
	Second E
	Third  E
}

func NewHTriple[E any](first E, second E, third E) HTriple[E] {
	return HTriple[E]{First: first, Second: second, Third: third}
}

func (ht HTriple[E]) Elements() []E {
	return []E{ht.First, ht.Second, ht.Third}
}

func (ht HTriple[E]) All() (E, E, E) {
	return ht.First, ht.Second, ht.Third
}

// Quadruple is a simple struct that holds four values.
type Quadruple[E1, E2, E3, E4 any] struct {
	First  E1
	Second E2
	Third  E3
	Fourth E4
}

func NewQuadruple[E1, E2, E3, E4 any](first E1, second E2, third E3, fourth E4) Quadruple[E1, E2, E3, E4] {
	return Quadruple[E1, E2, E3, E4]{First: first, Second: second, Third: third, Fourth: fourth}
}

func (q Quadruple[E1, E2, E3, E4]) Elements() []interface{} {
	return []interface{}{q.First, q.Second, q.Third, q.Fourth}
}

func (q Quadruple[E1, E2, E3, E4]) All() (E1, E2, E3, E4) {
	return q.First, q.Second, q.Third, q.Fourth
}

// HQuadruple is a homogenous struct that holds four values of the same type.
type HQuadruple[E any] struct {
	First  E
	Second E
	Third  E
	Fourth E
}

func NewHQuadruple[E any](first E, second E, third E, fourth E) HQuadruple[E] {
	return HQuadruple[E]{First: first, Second: second, Third: third, Fourth: fourth}
}

func (hq HQuadruple[E]) Elements() []E {
	return []E{hq.First, hq.Second, hq.Third, hq.Fourth}
}

func (hq HQuadruple[E]) All() (E, E, E, E) {
	return hq.First, hq.Second, hq.Third, hq.Fourth
}
