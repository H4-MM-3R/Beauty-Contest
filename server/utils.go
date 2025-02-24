package server

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateHubHash returns a random 8-character hex string.
func GenerateHubHash() (string, error) {
	bytes := make([]byte, 4) // 4 bytes => 8 hex characters.
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

/* Game Logic */

func defaultLogic(h *Hub, activeCount int) ([]string, float64, float64) {
	winners := []string{}
	sum := 0

	for client, val := range h.responses {
		if !client.eliminated {
			sum += val
		}
	}
	avg := float64(sum) / float64(activeCount)
	target := avg * 0.8

	// Determine winners: active players whose response is closest to the target.
	var minDiff float64 = 1e9
	for client, val := range h.responses {
		if client.eliminated {
			continue
		}
		diff := abs(float64(val) - target)
		if diff < minDiff {
			minDiff = diff
			winners = []string{client.name}
		} else if diff == minDiff {
			winners = append(winners, client.name)
		}
	}
	return winners, avg, target
}

func twoPlayerLogic(h *Hub) ([]string, float64, float64) {
	type Entry struct {
		client *Client
		Hand   int
	}

    var winMap = map[int]int {
        0b001: 0b101, // (rock vs scissors) => rock wins
        0b010: 0b011, // (paper vs rock) => paper wins
        0b100: 0b110, // (scissors vs paper) => scissors wins
    }

	entries := []Entry{}
	winners := []string{}
	for client, val := range h.responses {
		if client.eliminated {
			continue
		}
        hand := 0
        switch val {
        case 0: // rock
            hand = 0b001 // 1 (rock)
        case 100: // 
            hand = 0b010 // 2 (paper)
        default: 
            hand = 0b100 // 4 (scissors)
        }
		entries = append(entries, Entry{client, hand})
	}

    entry1 := entries[0]
    entry2 := entries[1]

    if entry1.Hand == entry2.Hand {
        winners = append(winners, entry1.client.name)
        winners = append(winners, entry2.client.name)
    } else if (entry1.Hand | entry2.Hand) == winMap[entry1.Hand] {
        winners = append(winners, entry1.client.name)
    } else {
        winners = append(winners, entry2.client.name)
    }

	return winners, -1, -1
}

func updateScores(h *Hub, winners []string) {

	// Update scores for each active player.
	for client := range h.clients {
		if client.eliminated {
			continue
		}
		isWinner := false
		for _, w := range winners {
			if w == client.name {
				isWinner = true
				break
			}
		}
		if !isWinner {
			client.score--
			if client.score <= 0 {
				client.eliminated = true
			}
		}
	}
}
