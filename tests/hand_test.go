package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	
	"github.com/primoPoker/server/pkg/poker"
)

func TestHandRoyalFlush(t *testing.T) {
	cards := []poker.Card{
		{poker.Ace, poker.Spades},
		{poker.King, poker.Spades},
		{poker.Queen, poker.Spades},
		{poker.Jack, poker.Spades},
		{poker.Ten, poker.Spades},
	}
	
	hand := poker.NewHand(cards)
	assert.Equal(t, poker.RoyalFlush, hand.Rank)
}

func TestHandStraightFlush(t *testing.T) {
	cards := []poker.Card{
		{poker.Nine, poker.Hearts},
		{poker.Eight, poker.Hearts},
		{poker.Seven, poker.Hearts},
		{poker.Six, poker.Hearts},
		{poker.Five, poker.Hearts},
	}
	
	hand := poker.NewHand(cards)
	assert.Equal(t, poker.StraightFlush, hand.Rank)
}

func TestHandFourOfAKind(t *testing.T) {
	cards := []poker.Card{
		{poker.Ace, poker.Spades},
		{poker.Ace, poker.Hearts},
		{poker.Ace, poker.Diamonds},
		{poker.Ace, poker.Clubs},
		{poker.King, poker.Spades},
	}
	
	hand := poker.NewHand(cards)
	assert.Equal(t, poker.FourOfAKind, hand.Rank)
}

func TestHandFullHouse(t *testing.T) {
	cards := []poker.Card{
		{poker.King, poker.Spades},
		{poker.King, poker.Hearts},
		{poker.King, poker.Diamonds},
		{poker.Ace, poker.Clubs},
		{poker.Ace, poker.Spades},
	}
	
	hand := poker.NewHand(cards)
	assert.Equal(t, poker.FullHouse, hand.Rank)
}

func TestHandFlush(t *testing.T) {
	cards := []poker.Card{
		{poker.Ace, poker.Spades},
		{poker.King, poker.Spades},
		{poker.Nine, poker.Spades},
		{poker.Seven, poker.Spades},
		{poker.Five, poker.Spades},
	}
	
	hand := poker.NewHand(cards)
	assert.Equal(t, poker.Flush, hand.Rank)
}

func TestHandStraight(t *testing.T) {
	cards := []poker.Card{
		{poker.Ace, poker.Spades},
		{poker.King, poker.Hearts},
		{poker.Queen, poker.Diamonds},
		{poker.Jack, poker.Clubs},
		{poker.Ten, poker.Spades},
	}
	
	hand := poker.NewHand(cards)
	assert.Equal(t, poker.Straight, hand.Rank)
}

func TestHandWheelStraight(t *testing.T) {
	// A-2-3-4-5 straight (wheel)
	cards := []poker.Card{
		{poker.Ace, poker.Spades},
		{poker.Five, poker.Hearts},
		{poker.Four, poker.Diamonds},
		{poker.Three, poker.Clubs},
		{poker.Two, poker.Spades},
	}
	
	hand := poker.NewHand(cards)
	assert.Equal(t, poker.Straight, hand.Rank)
	assert.Equal(t, poker.Five, hand.Kickers[0]) // 5 is high in wheel
}

func TestHandThreeOfAKind(t *testing.T) {
	cards := []poker.Card{
		{poker.King, poker.Spades},
		{poker.King, poker.Hearts},
		{poker.King, poker.Diamonds},
		{poker.Ace, poker.Clubs},
		{poker.Queen, poker.Spades},
	}
	
	hand := poker.NewHand(cards)
	assert.Equal(t, poker.ThreeOfAKind, hand.Rank)
}

func TestHandTwoPair(t *testing.T) {
	cards := []poker.Card{
		{poker.Ace, poker.Spades},
		{poker.Ace, poker.Hearts},
		{poker.King, poker.Diamonds},
		{poker.King, poker.Clubs},
		{poker.Queen, poker.Spades},
	}
	
	hand := poker.NewHand(cards)
	assert.Equal(t, poker.TwoPair, hand.Rank)
}

func TestHandOnePair(t *testing.T) {
	cards := []poker.Card{
		{poker.Ace, poker.Spades},
		{poker.Ace, poker.Hearts},
		{poker.King, poker.Diamonds},
		{poker.Queen, poker.Clubs},
		{poker.Jack, poker.Spades},
	}
	
	hand := poker.NewHand(cards)
	assert.Equal(t, poker.OnePair, hand.Rank)
}

func TestHandHighCard(t *testing.T) {
	cards := []poker.Card{
		{poker.Ace, poker.Spades},
		{poker.King, poker.Hearts},
		{poker.Queen, poker.Diamonds},
		{poker.Jack, poker.Clubs},
		{poker.Nine, poker.Spades},
	}
	
	hand := poker.NewHand(cards)
	assert.Equal(t, poker.HighCard, hand.Rank)
}

func TestCompareHands(t *testing.T) {
	// Royal flush vs straight flush
	royalFlush := poker.NewHand([]poker.Card{
		{poker.Ace, poker.Spades}, {poker.King, poker.Spades}, {poker.Queen, poker.Spades}, {poker.Jack, poker.Spades}, {poker.Ten, poker.Spades},
	})
	straightFlush := poker.NewHand([]poker.Card{
		{poker.Nine, poker.Hearts}, {poker.Eight, poker.Hearts}, {poker.Seven, poker.Hearts}, {poker.Six, poker.Hearts}, {poker.Five, poker.Hearts},
	})
	
	assert.Equal(t, 1, poker.CompareHands(royalFlush, straightFlush))
	assert.Equal(t, -1, poker.CompareHands(straightFlush, royalFlush))
	
	// Same rank hands
	pair1 := poker.NewHand([]poker.Card{
		{poker.Ace, poker.Spades}, {poker.Ace, poker.Hearts}, {poker.King, poker.Diamonds}, {poker.Queen, poker.Clubs}, {poker.Jack, poker.Spades},
	})
	pair2 := poker.NewHand([]poker.Card{
		{poker.King, poker.Spades}, {poker.King, poker.Hearts}, {poker.Ace, poker.Diamonds}, {poker.Queen, poker.Clubs}, {poker.Jack, poker.Spades},
	})
	
	assert.Equal(t, 1, poker.CompareHands(pair1, pair2)) // Pair of Aces > Pair of Kings
	
	// Identical hands
	pair3 := poker.NewHand([]poker.Card{
		{poker.Ace, poker.Clubs}, {poker.Ace, poker.Diamonds}, {poker.King, poker.Hearts}, {poker.Queen, poker.Spades}, {poker.Jack, poker.Hearts},
	})
	
	assert.Equal(t, 0, poker.CompareHands(pair1, pair3))
}

func TestGetBestHand(t *testing.T) {
	// 7 cards: 2 hole cards + 5 community cards
	cards := []poker.Card{
		{poker.Ace, poker.Spades},    // hole
		{poker.Ace, poker.Hearts},    // hole
		{poker.Ace, poker.Diamonds},  // community
		{poker.King, poker.Clubs},    // community
		{poker.King, poker.Spades},   // community
		{poker.Queen, poker.Hearts},  // community
		{poker.Jack, poker.Diamonds}, // community
	}
	
	bestHand := poker.GetBestHand(cards)
	assert.Equal(t, poker.FullHouse, bestHand.Rank) // AAA KK
}

func TestHandRankString(t *testing.T) {
	assert.Equal(t, "Royal Flush", poker.RoyalFlush.String())
	assert.Equal(t, "Straight Flush", poker.StraightFlush.String())
	assert.Equal(t, "Four of a Kind", poker.FourOfAKind.String())
	assert.Equal(t, "Full House", poker.FullHouse.String())
	assert.Equal(t, "Flush", poker.Flush.String())
	assert.Equal(t, "Straight", poker.Straight.String())
	assert.Equal(t, "Three of a Kind", poker.ThreeOfAKind.String())
	assert.Equal(t, "Two Pair", poker.TwoPair.String())
	assert.Equal(t, "One Pair", poker.OnePair.String())
	assert.Equal(t, "High Card", poker.HighCard.String())
}
