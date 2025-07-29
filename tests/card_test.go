package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/primoPoker/server/pkg/poker"
)

func TestNewDeck(t *testing.T) {
	deck := poker.NewDeck()
	
	// Should have 52 cards
	assert.Equal(t, 52, len(deck.Cards))
	
	// Should have 4 suits and 13 ranks
	suitCount := make(map[poker.Suit]int)
	rankCount := make(map[poker.Rank]int)
	
	for _, card := range deck.Cards {
		suitCount[card.Suit]++
		rankCount[card.Rank]++
	}
	
	// Each suit should appear 13 times
	for suit := poker.Hearts; suit <= poker.Spades; suit++ {
		assert.Equal(t, 13, suitCount[suit])
	}
	
	// Each rank should appear 4 times
	for rank := poker.Two; rank <= poker.Ace; rank++ {
		assert.Equal(t, 4, rankCount[rank])
	}
}

func TestDeckShuffle(t *testing.T) {
	deck1 := poker.NewDeck()
	deck2 := poker.NewDeck()
	
	// Save original order
	original := make([]poker.Card, len(deck1.Cards))
	copy(original, deck1.Cards)
	
	// Shuffle deck1
	deck1.Shuffle()
	
	// Shuffled deck should be different from original (with very high probability)
	different := false
	for i, card := range deck1.Cards {
		if card != original[i] {
			different = true
			break
		}
	}
	assert.True(t, different, "Shuffled deck should be different from original order")
	
	// Both decks should still have the same cards (just in different order)
	assert.ElementsMatch(t, deck1.Cards, deck2.Cards)
}

func TestDeckDeal(t *testing.T) {
	deck := poker.NewDeck()
	originalCount := len(deck.Cards)
	
	// Deal a card
	card, err := deck.Deal()
	require.NoError(t, err)
	
	// Should be a valid card
	assert.True(t, card.Rank >= poker.Two && card.Rank <= poker.Ace)
	assert.True(t, card.Suit >= poker.Hearts && card.Suit <= poker.Spades)
	
	// Deck should have one less card
	assert.Equal(t, originalCount-1, len(deck.Cards))
	
	// Deal all remaining cards
	for len(deck.Cards) > 0 {
		_, err := deck.Deal()
		require.NoError(t, err)
	}
	
	// Should not be able to deal from empty deck
	_, err = deck.Deal()
	assert.Error(t, err)
}

func TestDeckDealMultiple(t *testing.T) {
	deck := poker.NewDeck()
	
	// Deal 5 cards
	cards, err := deck.DealMultiple(5)
	require.NoError(t, err)
	assert.Equal(t, 5, len(cards))
	assert.Equal(t, 47, deck.Remaining())
	
	// Should not be able to deal more cards than available
	_, err = deck.DealMultiple(50)
	assert.Error(t, err)
}

func TestDeckReset(t *testing.T) {
	deck := poker.NewDeck()
	
	// Deal some cards
	deck.DealMultiple(10)
	assert.Equal(t, 42, deck.Remaining())
	
	// Reset deck
	deck.Reset()
	assert.Equal(t, 52, deck.Remaining())
	
	// Should be shuffled (different order than a new deck)
	newDeck := poker.NewDeck()
	different := false
	for i, card := range deck.Cards {
		if card != newDeck.Cards[i] {
			different = true
			break
		}
	}
	assert.True(t, different, "Reset deck should be shuffled")
}

func TestCardString(t *testing.T) {
	card := poker.NewCard(poker.Ace, poker.Spades)
	assert.Equal(t, "A♠", card.String())
	
	card = poker.NewCard(poker.Ten, poker.Hearts)
	assert.Equal(t, "10♥", card.String())
	
	card = poker.NewCard(poker.Two, poker.Diamonds)
	assert.Equal(t, "2♦", card.String())
}

func TestCardValue(t *testing.T) {
	assert.Equal(t, 2, poker.NewCard(poker.Two, poker.Hearts).Value())
	assert.Equal(t, 10, poker.NewCard(poker.Ten, poker.Hearts).Value())
	assert.Equal(t, 11, poker.NewCard(poker.Jack, poker.Hearts).Value())
	assert.Equal(t, 12, poker.NewCard(poker.Queen, poker.Hearts).Value())
	assert.Equal(t, 13, poker.NewCard(poker.King, poker.Hearts).Value())
	assert.Equal(t, 14, poker.NewCard(poker.Ace, poker.Hearts).Value())
}
