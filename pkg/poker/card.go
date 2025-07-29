package poker

import (
	"fmt"
	"math/rand"
	"time"
)

// Suit represents a card suit
type Suit int

const (
	Hearts Suit = iota
	Diamonds
	Clubs
	Spades
)

var suitNames = []string{"Hearts", "Diamonds", "Clubs", "Spades"}
var suitSymbols = []string{"♥", "♦", "♣", "♠"}

func (s Suit) String() string {
	return suitNames[s]
}

func (s Suit) Symbol() string {
	return suitSymbols[s]
}

// Rank represents a card rank
type Rank int

const (
	Two Rank = iota + 2
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

var rankNames = []string{"", "", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}

func (r Rank) String() string {
	return rankNames[r]
}

// Card represents a playing card
type Card struct {
	Rank Rank `json:"rank"`
	Suit Suit `json:"suit"`
}

// NewCard creates a new card
func NewCard(rank Rank, suit Suit) Card {
	return Card{Rank: rank, Suit: suit}
}

// String returns a string representation of the card
func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Rank, c.Suit.Symbol())
}

// Value returns the numerical value of the card for comparison
func (c Card) Value() int {
	return int(c.Rank)
}

// Deck represents a deck of cards
type Deck struct {
	Cards []Card `json:"cards"`
	rng   *rand.Rand
}

// NewDeck creates a new standard 52-card deck
func NewDeck() *Deck {
	deck := &Deck{
		Cards: make([]Card, 0, 52),
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Create all 52 cards
	for suit := Hearts; suit <= Spades; suit++ {
		for rank := Two; rank <= Ace; rank++ {
			deck.Cards = append(deck.Cards, NewCard(rank, suit))
		}
	}

	return deck
}

// Shuffle shuffles the deck using Fisher-Yates algorithm
func (d *Deck) Shuffle() {
	for i := len(d.Cards) - 1; i > 0; i-- {
		j := d.rng.Intn(i + 1)
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	}
}

// Deal deals the top card from the deck
func (d *Deck) Deal() (Card, error) {
	if len(d.Cards) == 0 {
		return Card{}, fmt.Errorf("cannot deal from empty deck")
	}

	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	return card, nil
}

// DealMultiple deals multiple cards from the deck
func (d *Deck) DealMultiple(count int) ([]Card, error) {
	if len(d.Cards) < count {
		return nil, fmt.Errorf("not enough cards in deck: have %d, need %d", len(d.Cards), count)
	}

	cards := make([]Card, count)
	for i := 0; i < count; i++ {
		card, err := d.Deal()
		if err != nil {
			return nil, err
		}
		cards[i] = card
	}

	return cards, nil
}

// Remaining returns the number of cards remaining in the deck
func (d *Deck) Remaining() int {
	return len(d.Cards)
}

// Reset resets the deck to a full 52-card deck and shuffles it
func (d *Deck) Reset() {
	d.Cards = d.Cards[:0]
	
	// Recreate all 52 cards
	for suit := Hearts; suit <= Spades; suit++ {
		for rank := Two; rank <= Ace; rank++ {
			d.Cards = append(d.Cards, NewCard(rank, suit))
		}
	}
	
	d.Shuffle()
}
