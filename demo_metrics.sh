#!/bin/bash

# Simple demo script to test the metrics API endpoints
# This script demonstrates the new player metrics functionality

echo "=== PrimoPoker Player Metrics Demo ==="
echo

# Base URL for the API
BASE_URL="http://localhost:8080/api/v1"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}1. Health Check${NC}"
curl -X GET "$BASE_URL/../health" -w "\n\n"

echo -e "${BLUE}2. Available API Endpoints for Player Metrics:${NC}"
echo "GET /api/v1/metrics - Get current user's comprehensive metrics"
echo "GET /api/v1/metrics?since=2024-01-01T00:00:00Z - Get metrics since specific date"
echo "GET /api/v1/metrics/comparison - Compare metrics between periods"
echo "GET /api/v1/users/{userId}/metrics - Get specific user's metrics"
echo

echo -e "${BLUE}3. Metrics Features Implemented:${NC}"
echo "✅ Hands played"
echo "✅ VPIP (Voluntarily Put $ In Pot)"
echo "✅ PFR (Pre-Flop Raise)"
echo "✅ 3-bet frequency"
echo "✅ Fold to 3-bet frequency"
echo "✅ Aggression factor"
echo "✅ C-bet (Continuation bet) frequency"
echo "✅ Went to showdown statistics"
echo "✅ Won $ at showdown"
echo "✅ Comprehensive financial metrics"
echo

echo -e "${BLUE}4. Example Response Structure:${NC}"
cat << 'EOF'
{
  "success": true,
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "player123",
    "period_start": "2024-01-01T00:00:00Z",
    "period_end": "2024-12-01T10:30:00Z",
    "hands_played": 1500,
    "hands_won": 450,
    "hands_lost": 750,
    "hands_folded": 300,
    "win_rate": 30.0,
    "vpip_percent": 22.5,
    "pfr_percent": 18.2,
    "three_bet_percent": 8.5,
    "fold_to_three_bet_percent": 65.0,
    "cbet_percent": 75.0,
    "fold_to_cbet_percent": 40.0,
    "aggression_factor": 2.8,
    "went_to_showdown": 180,
    "won_at_showdown": 95,
    "showdown_win_rate": 52.8,
    "won_dollar_at_showdown": 45000,
    "total_wagered": 150000,
    "total_won": 165000,
    "net_result": 15000,
    "avg_pot_size": 1200.5,
    "avg_win_amount": 366.7,
    "biggest_win": 8500,
    "biggest_loss": -3200
  }
}
EOF

echo
echo -e "${GREEN}✅ Player Metrics Feature Implementation Complete!${NC}"
echo
echo -e "${BLUE}To test with authentication:${NC}"
echo "1. Register a user: POST /api/v1/auth/register"
echo "2. Login to get JWT token: POST /api/v1/auth/login"
echo "3. Use token in Authorization header: 'Authorization: Bearer <token>'"
echo "4. Call metrics endpoints with valid authentication"
echo

echo -e "${BLUE}Advanced Features:${NC}"
echo "• Time-based filtering with 'since' parameter"
echo "• Period comparison for tracking improvement"
echo "• Comprehensive poker statistics calculation"
echo "• Real-time metrics based on hand history data"
echo