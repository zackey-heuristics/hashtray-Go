package permutator

import (
	"context"
	"math"
)

// Permutator generates all email address combinations from chunks and domains.
type Permutator struct {
	Chunks     []string
	Domains    []string
	Separators []string
	Crazy      bool
}

// New creates a Permutator with default separators.
func New(chunks, domains []string, crazy bool) *Permutator {
	return &Permutator{
		Chunks:     chunks,
		Domains:    domains,
		Separators: []string{"", ".", "_", "-"},
		Crazy:      crazy,
	}
}

// CombinationCount returns the total number of email combinations that will be generated.
// Returns -1 if the count would overflow int.
func (p *Permutator) CombinationCount() int {
	n := len(p.Chunks)
	if n == 0 || len(p.Domains) == 0 {
		return 0
	}
	total := 0
	for r := 1; r <= n; r++ {
		var count int
		if r == 1 {
			count = n
		} else {
			combCount := combinations(n, r)
			permCount := factorial(r)
			if combCount < 0 || permCount < 0 {
				return -1
			}
			count = safeMul(combCount, permCount)
			if count < 0 {
				return -1
			}
			if p.Crazy {
				count = safeMul(count, intPow(len(p.Separators), r-1))
			} else {
				count = safeMul(count, len(p.Separators))
			}
			if count < 0 {
				return -1
			}
		}
		total = safeAdd(total, count)
		if total < 0 {
			return -1
		}
	}
	total = safeMul(total, len(p.Domains))
	if total < 0 {
		return -1
	}
	return total
}

// GenerateCtx returns a channel that yields all email combinations lazily.
// Uses backtracking to avoid materializing all permutations in memory.
// The goroutine stops when ctx is cancelled, preventing goroutine leaks.
func (p *Permutator) GenerateCtx(ctx context.Context) <-chan string {
	ch := make(chan string, 256)
	go func() {
		defer close(ch)
		n := len(p.Chunks)
		if n == 0 || len(p.Domains) == 0 {
			return
		}

		send := func(address string) bool {
			select {
			case ch <- address:
				return true
			case <-ctx.Done():
				return false
			}
		}

		for r := 1; r <= n; r++ {
			// Precompute separator combinations for this r (crazy mode)
			var sepCombos [][]string
			if r > 1 && p.Crazy {
				sepCombos = separatorProduct(p.Separators, r-1)
			}

			emitForPerm := func(perm []string) bool {
				for _, domain := range p.Domains {
					if len(perm) == 1 {
						if !send(perm[0] + "@" + domain) {
							return false
						}
					} else if p.Crazy {
						for _, seps := range sepCombos {
							if !send(interleave(perm, seps) + "@" + domain) {
								return false
							}
						}
					} else {
						for _, sep := range p.Separators {
							if !send(joinWith(perm, sep) + "@" + domain) {
								return false
							}
						}
					}
				}
				return true
			}

			// Lazy backtracking permutation generation
			current := make([]string, r)
			used := make([]bool, n)

			var backtrack func(pos int) bool
			backtrack = func(pos int) bool {
				select {
				case <-ctx.Done():
					return false
				default:
				}

				if pos == r {
					return emitForPerm(current)
				}

				for i := 0; i < n; i++ {
					if used[i] {
						continue
					}
					used[i] = true
					current[pos] = p.Chunks[i]
					if !backtrack(pos + 1) {
						return false
					}
					used[i] = false
				}
				return true
			}

			if !backtrack(0) {
				return
			}
		}
	}()
	return ch
}

// Generate returns a channel that yields all email combinations.
func (p *Permutator) Generate() <-chan string {
	return p.GenerateCtx(context.Background())
}

// separatorProduct returns the Cartesian product of separators repeated n times.
func separatorProduct(seps []string, n int) [][]string {
	if n == 0 {
		return [][]string{{}}
	}
	sub := separatorProduct(seps, n-1)
	var result [][]string
	for _, s := range seps {
		for _, prev := range sub {
			combo := make([]string, 0, n)
			combo = append(combo, s)
			combo = append(combo, prev...)
			result = append(result, combo)
		}
	}
	return result
}

// interleave joins elements with the corresponding separators: e0 s0 e1 s1 e2 ...
func interleave(elements, separators []string) string {
	var b []byte
	for i, e := range elements {
		b = append(b, e...)
		if i < len(separators) {
			b = append(b, separators[i]...)
		}
	}
	return string(b)
}

// joinWith joins all elements with a single separator.
func joinWith(elements []string, sep string) string {
	var b []byte
	for i, e := range elements {
		if i > 0 {
			b = append(b, sep...)
		}
		b = append(b, e...)
	}
	return string(b)
}

func factorial(n int) int {
	result := 1
	for i := 2; i <= n; i++ {
		result = safeMul(result, i)
		if result < 0 {
			return -1
		}
	}
	return result
}

func combinations(n, r int) int {
	if r > n {
		return 0
	}
	num := 1
	den := 1
	for i := 0; i < r; i++ {
		num = safeMul(num, n-i)
		den = safeMul(den, i+1)
		if num < 0 || den < 0 {
			return -1
		}
	}
	return num / den
}

func intPow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result = safeMul(result, base)
		if result < 0 {
			return -1
		}
	}
	return result
}

func safeMul(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	result := a * b
	if result/a != b || result > math.MaxInt/2 {
		return -1
	}
	return result
}

func safeAdd(a, b int) int {
	result := a + b
	if result < a || result > math.MaxInt/2 {
		return -1
	}
	return result
}
