package permutator

import "context"

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
func (p *Permutator) CombinationCount() int {
	n := len(p.Chunks)
	if n == 0 || len(p.Domains) == 0 {
		return 0
	}
	total := 0
	for r := 1; r <= n; r++ {
		if r == 1 {
			total += n
		} else {
			combCount := combinations(n, r)
			permCount := factorial(r)
			count := combCount * permCount
			if p.Crazy {
				count *= intPow(len(p.Separators), r-1)
			} else {
				count *= len(p.Separators)
			}
			total += count
		}
	}
	return total * len(p.Domains)
}

// GenerateCtx returns a channel that yields all email combinations.
// The goroutine stops when ctx is cancelled, preventing goroutine leaks.
func (p *Permutator) GenerateCtx(ctx context.Context) <-chan string {
	ch := make(chan string, 256)
	go func() {
		defer close(ch)
		n := len(p.Chunks)
		if n == 0 || len(p.Domains) == 0 {
			return
		}
		for r := 1; r <= n; r++ {
			perms := permutations(p.Chunks, r)
			for _, perm := range perms {
				for _, domain := range p.Domains {
					if len(perm) == 1 {
						select {
						case ch <- perm[0] + "@" + domain:
						case <-ctx.Done():
							return
						}
					} else if p.Crazy {
						for _, seps := range separatorProduct(p.Separators, r-1) {
							local := interleave(perm, seps)
							select {
							case ch <- local + "@" + domain:
							case <-ctx.Done():
								return
							}
						}
					} else {
						for _, sep := range p.Separators {
							local := joinWith(perm, sep)
							select {
							case ch <- local + "@" + domain:
							case <-ctx.Done():
								return
							}
						}
					}
				}
			}
		}
	}()
	return ch
}

// Generate returns a channel that yields all email combinations.
// Deprecated: Use GenerateCtx with a cancellable context to avoid goroutine leaks.
func (p *Permutator) Generate() <-chan string {
	return p.GenerateCtx(context.Background())
}

// permutations returns all ordered arrangements of length r from items.
func permutations(items []string, r int) [][]string {
	if r == 0 {
		return [][]string{{}}
	}
	var result [][]string
	for i, item := range items {
		remaining := make([]string, 0, len(items)-1)
		remaining = append(remaining, items[:i]...)
		remaining = append(remaining, items[i+1:]...)
		for _, sub := range permutations(remaining, r-1) {
			perm := make([]string, 0, r)
			perm = append(perm, item)
			perm = append(perm, sub...)
			result = append(result, perm)
		}
	}
	return result
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
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func combinations(n, r int) int {
	return factorial(n) / (factorial(r) * factorial(n-r))
}

func intPow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}
