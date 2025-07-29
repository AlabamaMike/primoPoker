package poker

import (
	"sort"
)

// HandRank represents the rank of a poker hand
type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

var handRankNames = []string{
	"High Card",
	"One Pair",
	"Two Pair",
	"Three of a Kind",
	"Straight",
	"Flush",
	"Full House",
	"Four of a Kind",
	"Straight Flush",
	"Royal Flush",
}

func (hr HandRank) String() string {
	return handRankNames[hr]
}

// Hand represents a poker hand
type Hand struct {
	Cards    []Card   `json:"cards"`
	Rank     HandRank `json:"rank"`
	Kickers  []Rank   `json:"kickers"`  // For tie-breaking
	Value    int      `json:"value"`    // Overall hand value for comparison
}

// NewHand creates a new hand from the given cards
func NewHand(cards []Card) *Hand {
	if len(cards) != 5 {
		panic("Hand must contain exactly 5 cards")
	}

	hand := &Hand{
		Cards: make([]Card, 5),
	}
	copy(hand.Cards, cards)

	// Sort cards by rank (descending)
	sort.Slice(hand.Cards, func(i, j int) bool {
		return hand.Cards[i].Rank > hand.Cards[j].Rank
	})

	hand.evaluate()
	return hand
}

// evaluate determines the rank and value of the hand
func (h *Hand) evaluate() {
	ranks := make([]int, 15) // Index by rank (2-14)
	suits := make([]int, 4)  // Index by suit (0-3)

	// Count ranks and suits
	for _, card := range h.Cards {
		ranks[card.Rank]++
		suits[card.Suit]++
	}

	// Check for flush
	isFlush := false
	for _, count := range suits {
		if count == 5 {
			isFlush = true
			break
		}
	}

	// Check for straight
	isStraight, straightHigh := h.checkStraight(ranks)

	// Determine hand rank
	if isStraight && isFlush {
		if straightHigh == Ace && h.Cards[1].Rank == King {
			h.Rank = RoyalFlush
			h.Value = int(RoyalFlush)*100000000 + int(Ace)
		} else {
			h.Rank = StraightFlush
			h.Value = int(StraightFlush)*100000000 + int(straightHigh)
		}
		return
	}

	// Count pairs, trips, quads
	pairs := []Rank{}
	trips := []Rank{}
	quads := []Rank{}
	kickers := []Rank{}

	for rank := Ace; rank >= Two; rank-- {
		count := ranks[rank]
		switch count {
		case 4:
			quads = append(quads, rank)
		case 3:
			trips = append(trips, rank)
		case 2:
			pairs = append(pairs, rank)
		case 1:
			kickers = append(kickers, rank)
		}
	}

	// Determine hand rank based on pairs, trips, quads
	if len(quads) == 1 {
		h.Rank = FourOfAKind
		h.Kickers = append([]Rank{quads[0]}, kickers...)
		h.Value = int(FourOfAKind)*100000000 + int(quads[0])*1000000 + int(kickers[0])
	} else if len(trips) == 1 && len(pairs) == 1 {
		h.Rank = FullHouse
		h.Kickers = []Rank{trips[0], pairs[0]}
		h.Value = int(FullHouse)*100000000 + int(trips[0])*1000000 + int(pairs[0])
	} else if isFlush {
		h.Rank = Flush
		h.Kickers = kickers
		h.Value = int(Flush)*100000000 + h.kickerValue(kickers)
	} else if isStraight {
		h.Rank = Straight
		h.Kickers = []Rank{straightHigh}
		h.Value = int(Straight)*100000000 + int(straightHigh)
	} else if len(trips) == 1 {
		h.Rank = ThreeOfAKind
		h.Kickers = append([]Rank{trips[0]}, kickers...)
		h.Value = int(ThreeOfAKind)*100000000 + int(trips[0])*1000000 + h.kickerValue(kickers)
	} else if len(pairs) == 2 {
		h.Rank = TwoPair
		h.Kickers = append(pairs, kickers...)
		h.Value = int(TwoPair)*100000000 + int(pairs[0])*1000000 + int(pairs[1])*10000 + int(kickers[0])
	} else if len(pairs) == 1 {
		h.Rank = OnePair
		h.Kickers = append([]Rank{pairs[0]}, kickers...)
		h.Value = int(OnePair)*100000000 + int(pairs[0])*1000000 + h.kickerValue(kickers)
	} else {
		h.Rank = HighCard
		h.Kickers = kickers
		h.Value = int(HighCard)*100000000 + h.kickerValue(kickers)
	}
}

// checkStraight checks if the hand contains a straight
func (h *Hand) checkStraight(ranks []int) (bool, Rank) {
	// Check for normal straight
	consecutive := 0
	var high Rank
	
	for rank := Ace; rank >= Two; rank-- {
		if ranks[rank] > 0 {
			consecutive++
			if consecutive == 1 {
				high = rank
			}
		} else {
			consecutive = 0
		}
		
		if consecutive == 5 {
			return true, high
		}
	}
	
	// Check for A-2-3-4-5 straight (wheel)
	if ranks[Ace] > 0 && ranks[Two] > 0 && ranks[Three] > 0 && ranks[Four] > 0 && ranks[Five] > 0 {
		return true, Five // In a wheel, the 5 is the high card
	}
	
	return false, 0
}

// kickerValue calculates a numeric value for kicker comparison
func (h *Hand) kickerValue(kickers []Rank) int {
	value := 0
	multiplier := 1
	
	for i := len(kickers) - 1; i >= 0; i-- {
		value += int(kickers[i]) * multiplier
		multiplier *= 100
	}
	
	return value
}

// Compare compares two hands, returns 1 if h1 wins, -1 if h2 wins, 0 for tie
func CompareHands(h1, h2 *Hand) int {
	if h1.Value > h2.Value {
		return 1
	} else if h1.Value < h2.Value {
		return -1
	}
	return 0
}

// GetBestHand finds the best 5-card hand from 7 cards (2 hole + 5 community)
func GetBestHand(cards []Card) *Hand {
	if len(cards) != 7 {
		panic("Must provide exactly 7 cards")
	}

	var bestHand *Hand
	
	// Generate all possible 5-card combinations from 7 cards
	combinations := generateCombinations(cards, 5)
	
	for _, combo := range combinations {
		hand := NewHand(combo)
		if bestHand == nil || CompareHands(hand, bestHand) > 0 {
			bestHand = hand
		}
	}
	
	return bestHand
}

// generateCombinations generates all combinations of r items from a slice
func generateCombinations(cards []Card, r int) [][]Card {
	var result [][]Card
	
	var backtrack func(start int, current []Card)
	backtrack = func(start int, current []Card) {
		if len(current) == r {
			combo := make([]Card, r)
			copy(combo, current)
			result = append(result, combo)
			return
		}
		
		for i := start; i < len(cards); i++ {
			current = append(current, cards[i])
			backtrack(i+1, current)
			current = current[:len(current)-1]
		}
	}
	
	backtrack(0, []Card{})
	return result
}
