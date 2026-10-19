package host

import (
	"errors"
	"unicode/utf8"
)

var errBadGlobPattern = errors.New("syntax error in pattern")

type globTokenKind uint8

const (
	globLiteral globTokenKind = iota
	globAny
	globStar
	globClass
)

type globRange struct {
	lo rune
	hi rune
}

type globToken struct {
	kind    globTokenKind
	literal byte
	negated bool
	ranges  []globRange
}

type globPattern struct {
	tokens []globToken
}

func compileGlobPattern(pattern string) (*globPattern, error) {
	compiled := &globPattern{}
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '*':
			if len(compiled.tokens) == 0 || compiled.tokens[len(compiled.tokens)-1].kind != globStar {
				compiled.tokens = append(compiled.tokens, globToken{kind: globStar})
			}
			i++
		case '?':
			compiled.tokens = append(compiled.tokens, globToken{kind: globAny})
			i++
		case '[':
			token, next, err := parseGlobClass(pattern, i+1)
			if err != nil {
				return nil, err
			}
			compiled.tokens = append(compiled.tokens, token)
			i = next
		case '\\':
			i++
			if i == len(pattern) {
				return nil, errBadGlobPattern
			}
			compiled.tokens = append(compiled.tokens, globToken{kind: globLiteral, literal: pattern[i]})
			i++
		default:
			compiled.tokens = append(compiled.tokens, globToken{kind: globLiteral, literal: pattern[i]})
			i++
		}
	}
	return compiled, nil
}

func parseGlobClass(pattern string, i int) (globToken, int, error) {
	token := globToken{kind: globClass}
	if i < len(pattern) && pattern[i] == '^' {
		token.negated = true
		i++
	}

	for {
		if i < len(pattern) && pattern[i] == ']' && len(token.ranges) > 0 {
			return token, i + 1, nil
		}

		lo, next, err := readGlobClassRune(pattern, i)
		if err != nil {
			return globToken{}, 0, err
		}
		hi := lo
		i = next
		if i < len(pattern) && pattern[i] == '-' {
			hi, i, err = readGlobClassRune(pattern, i+1)
			if err != nil {
				return globToken{}, 0, err
			}
		}
		token.ranges = append(token.ranges, globRange{lo: lo, hi: hi})
	}
}

func readGlobClassRune(pattern string, i int) (rune, int, error) {
	if i >= len(pattern) || pattern[i] == '-' || pattern[i] == ']' {
		return 0, 0, errBadGlobPattern
	}
	if pattern[i] == '\\' {
		i++
		if i >= len(pattern) {
			return 0, 0, errBadGlobPattern
		}
	}
	r, size := utf8.DecodeRuneInString(pattern[i:])
	if r == utf8.RuneError && size == 1 {
		return 0, 0, errBadGlobPattern
	}
	next := i + size
	if next == len(pattern) {
		return 0, 0, errBadGlobPattern
	}
	return r, next, nil
}

func (p *globPattern) Match(name string) bool {
	tokens := p.tokens
	tokenIndex, nameIndex := 0, 0
	starIndex, starNameIndex := -1, -1

	for nameIndex < len(name) {
		if tokenIndex < len(tokens) && tokens[tokenIndex].kind != globStar {
			if size, ok := tokens[tokenIndex].matchAt(name, nameIndex); ok {
				tokenIndex++
				nameIndex += size
				continue
			}
		}

		if tokenIndex < len(tokens) && tokens[tokenIndex].kind == globStar {
			starIndex = tokenIndex
			starNameIndex = nameIndex
			tokenIndex++
			continue
		}

		if starIndex < 0 || starNameIndex >= len(name) || name[starNameIndex] == '/' {
			return false
		}
		starNameIndex++
		nameIndex = starNameIndex
		tokenIndex = starIndex + 1
	}

	for tokenIndex < len(tokens) && tokens[tokenIndex].kind == globStar {
		tokenIndex++
	}
	return tokenIndex == len(tokens)
}

func (t globToken) matchAt(name string, i int) (int, bool) {
	switch t.kind {
	case globLiteral:
		return 1, name[i] == t.literal
	case globAny:
		if name[i] == '/' {
			return 1, false
		}
		_, size := utf8.DecodeRuneInString(name[i:])
		return size, true
	case globClass:
		if name[i] == '/' {
			return 1, false
		}
		r, size := utf8.DecodeRuneInString(name[i:])
		matched := false
		for _, candidate := range t.ranges {
			if candidate.lo <= r && r <= candidate.hi {
				matched = true
				break
			}
		}
		return size, matched != t.negated
	default:
		return 0, false
	}
}
