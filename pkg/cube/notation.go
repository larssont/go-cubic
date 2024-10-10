package cube

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/zyedidia/generic/stack"
)

// Regex based on official WCA notation https://www.worldcubeassociation.org/regulations/#article-12-notation
// Also includes support for commutators and conjugates
var reToken = regexp.MustCompile(
	`^` +
		`(?P<slices>\d?)` + // Optional digit for number of slices
		`(?P<face>[ULFRBDMESxyz])` + // A single character for the move (U, L, F, R, B, D, M, E, S, x, y, z)
		`(?P<wide>w?)` + // Optional 'w' for wide
		`(?P<rotations>\d?)` + // Optional digit for rotations
		`(?P<prime>'?)` + // Optional "'" for prime
		`(?P<end>[ \]\),:]?)`, // Optional end character (space, ']', ')', ',', ':')
)

type GroupType int

const (
	groupTypeNil GroupType = iota
	GroupTypeMove
	GroupTypeComm
)

var (
	OpenGroupTypes = map[byte]GroupType{
		'(': GroupTypeMove,
		'[': GroupTypeComm,
	}
	CloseGroupTypes = map[byte]GroupType{
		')': GroupTypeMove,
		']': GroupTypeComm,
	}

	ErrUnexpectedGroupClosure = errors.New("unexpected group closure")
	ErrUnclosedGroup          = errors.New("unclosed group")
	ErrTokenExtraction        = errors.New("token extraction")
	ErrRotationMove           = errors.New("slices or wide with x, y, z")
	ErrSlicesMove             = errors.New("slices without wide move")
	ErrMultipleSeparators     = errors.New("multiple separators in one group")
	ErrSeparatorGroup         = errors.New("separators not in comm group")
)

type Tokenizable interface {
	String() string
	CombineMove(Move) (*Move, bool)
}

type Group struct {
	Tokens    []Tokenizable
	Factor    int
	GroupType GroupType
}

func NewGroup(groupType GroupType) Group {
	return Group{
		GroupType: groupType,
		Factor:    1,
	}
}

func (t *Group) String() string {
	return fmt.Sprintf("Group: n=%d factor=%d type=%d", len(t.Tokens), t.Factor, t.GroupType)
}

func (t *Group) CombineMove(combo Move) (*Move, bool) {
	return nil, false
}

func (g *Group) AddToken(t Tokenizable) {
	g.Tokens = append(g.Tokens, t)
}

func (g *Group) Print() {
	g.print(0)
}

func (g *Group) print(level int) {
	indent := strings.Repeat("  ", level)
	fmt.Printf("%s%s\n", indent, g.String())

	for _, token := range g.Tokens {
		group, ok := token.(*Group)
		if ok {
			group.print(level + 1)
		} else {
			fmt.Printf("%s%s\n", indent, token.String())
		}

	}
}

// Expand returns a flattened slice of Moves from the Group, handling nested groups and separators.
// Moves are repeated by if the group Factor > 1.
// Returns an error for multiple or invalid separators.
func (g *Group) Expand() ([]Move, error) {
	var head []Move
	var tail []Move
	var separator *Separator

	for _, token := range g.Tokens {
		switch v := token.(type) {
		case *Move:
			if separator == nil {
				head = append(head, *v)
			} else {
				tail = append(tail, *v)
			}
		case *Separator:
			if separator != nil {
				return nil, ErrMultipleSeparators
			}
			separator = v
		case *Group:
			tokens, err := v.Expand()
			if err != nil {
				return nil, err
			}

			if separator == nil {
				head = append(head, tokens...)
			} else {
				tail = append(tail, tokens...)
			}
		}
	}

	res := append(head, tail...)

	if separator != nil && g.GroupType != GroupTypeComm {
		return nil, ErrSeparatorGroup
	}

	// Conjugates have reversed head after tail
	if separator == &SepConjugate {
		res = append(res, ReverseMoves(head)...)
	}

	// Commutators have reversed head and tail after normal tail
	if separator == &SepCommutator {
		res = append(res, ReverseMoves(head)...)
		res = append(res, ReverseMoves(tail)...)
	}

	out := res
	if g.Factor > 1 {
		out = make([]Move, 0, len(res)*g.Factor)
		for i := 0; i < g.Factor; i++ {
			out = append(out, res...)
		}
	}

	return NormalizeMoves(out)
}

func ReverseMoves(moves []Move) []Move {
	out := reverse(moves)

	for i, m := range out {
		newMove := m
		newMove.Inverted = !m.Inverted

		out[i] = newMove
	}

	return out
}

func NormalizeMoves(moves []Move) ([]Move, error) {
	normalized := make([]Move, 0, len(moves))

	for _, t := range moves {
		if len(normalized) == 0 {
			normalized = append(normalized, t)
			continue
		}

		lastIndex := len(normalized) - 1
		lastMove := normalized[lastIndex]

		combined, ok := lastMove.CombineMove(t)
		if !ok {
			normalized = append(normalized, t)
			continue
		}
		if combined == nil {
			normalized = normalized[:lastIndex]
			continue
		}

		combined.Normalize()
		normalized[lastIndex] = *combined
	}

	return normalized, nil
}

var (
	Separators = map[byte]*Separator{
		':': &SepConjugate,
		',': &SepCommutator,
	}

	SepCommutator = Separator{','}
	SepConjugate  = Separator{':'}
)

type Separator struct {
	Separator rune
}

func (s *Separator) String() string {
	return fmt.Sprintf("Separator: %s", string(s.Separator))
}

func (t *Separator) CombineMove(combo Move) (*Move, bool) {
	return nil, false
}

type Move struct {
	Slices    int
	Operator  rune
	Wide      bool
	Rotations int
	Inverted  bool
}

func (t *Move) String() string {
	return fmt.Sprintf(
		"Move: Slices=%d, Operator=%c, Wide=%t, Rotations=%d, Inverted=%t",
		t.Slices, t.Operator, t.Wide, t.Rotations, t.Inverted,
	)
}

// CombineMove merges two compatible moves.
// Returns the combined move or nil if they cancel out, and a bool indicating success.
func (t *Move) CombineMove(combo Move) (*Move, bool) {
	if t.Operator != combo.Operator || t.Slices != combo.Slices || t.Wide != combo.Wide {
		return nil, false
	}

	out := *t
	if out.Inverted == combo.Inverted {
		out.Rotations += combo.Rotations
	} else {
		out.Rotations -= combo.Rotations
	}

	if out.Rotations%4 == 0 {
		return nil, true
	}

	return &out, true
}

func (t *Move) isAny(runes ...rune) bool {
	return slices.Contains(runes, t.Operator)
}

func (t *Move) validate() error {
	if t.isAny('x', 'y', 'z') && (t.Slices > 0 || t.Wide) {
		return ErrRotationMove
	}

	if t.Slices != 0 && !t.Wide {
		return ErrSlicesMove
	}
	return nil
}

func (t *Move) Normalize() {
	if t.Slices == 0 && t.Wide {
		t.Slices = 2
	}

	t.Rotations = t.Rotations % 4

	if t.Rotations == 0 {
		t.Rotations = 1
	} else if t.Rotations == 3 {
		t.Rotations = 1
		t.Inverted = !t.Inverted
	} else if t.Rotations < 0 {
		t.Rotations *= -1
		t.Inverted = !t.Inverted
	}
}

func extractToken(input string) (*Move, int, error) {
	matches := reToken.FindStringSubmatch(input)

	if len(matches) > 0 {
		slices, err := strToInt(matches[reToken.SubexpIndex("slices")])
		if err != nil {
			return nil, 0, err
		}

		rotations, err := strToInt(matches[reToken.SubexpIndex("rotations")])
		if err != nil {
			return nil, 0, err
		}

		t := &Move{
			Slices:    slices,
			Operator:  rune(matches[reToken.SubexpIndex("face")][0]),
			Wide:      matches[reToken.SubexpIndex("wide")] == "w",
			Rotations: rotations,
			Inverted:  matches[reToken.SubexpIndex("prime")] == "'",
		}

		length := len(matches[0]) - len(matches[reToken.SubexpIndex("end")])
		t.Normalize()
		err = t.validate()

		return t, length, err
	}

	return nil, 0, ErrTokenExtraction
}

func ParseNotation(input string) (*Group, error) {
	defer timeTrack(time.Now(), "parse notation")

	stack := stack.New[Group]()
	currentGroup := NewGroup(groupTypeNil)

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if ch == ' ' {
			continue
		}

		if v, ok := OpenGroupTypes[ch]; ok {
			stack.Push(currentGroup)
			currentGroup = NewGroup(v)
		} else if _, ok := CloseGroupTypes[ch]; ok {
			group := currentGroup

			if stack.Size() == 0 {
				return nil, ErrUnexpectedGroupClosure
			}
			currentGroup = stack.Pop()

			i++
			if i < len(input) && isDigit(input[i]) {
				group.Factor = int(input[i] - '0')
				i++
			}

			currentGroup.AddToken(&group)
			i--
		} else if v, ok := Separators[ch]; ok {
			currentGroup.AddToken(v)
		} else {
			token, end, err := extractToken(input[i:])
			if err != nil {
				return nil, err
			}

			currentGroup.AddToken(token)
			i += end - 1
		}
	}

	if stack.Size() > 0 {
		return nil, ErrUnclosedGroup
	}

	return &currentGroup, nil
}
