package game

import (
	"bufio"
	"fmt"
	"math"
	"os"
)

// Beauty Contest (keynesian beauty contest) with 0.8 * avg score

func InitGame() (Game, *bufio.Scanner) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Enter the number of players: \n")
	playerCount := ConstraintScanner(scanner, 1, 7)
	ClearScreen()

	players := make([]Player, playerCount)
	for i := 0; i < playerCount; i++ {
		fmt.Printf("Enter player %d name: \n", i+1)
		scanner.Scan()
		players[i].Name = scanner.Text()
	}
	ClearScreen()

	fmt.Printf("Enter the elimination score: \n")
	eliminationScore := ConstraintScanner(scanner, 0, 10)
	ClearScreen()

	return Game{players, eliminationScore}, scanner
}

func PlayGame(game Game, scanner *bufio.Scanner) {
	round := 1
	for {
		isEliminated, eliminatedPlayerIdx := CheckPlayerElimination(game)
		fmt.Printf("===== Round %d start ======\n", round)
		PlayRound(game, round, scanner)
		if isEliminated && len(game.Players) > 1 {
            ClearScreen()
			fmt.Printf("Player %s has been eliminated\n", game.Players[eliminatedPlayerIdx].Name)
			game.Players = append(game.Players[:eliminatedPlayerIdx], game.Players[eliminatedPlayerIdx+1:]...)
            if len(game.Players) == 1 {
                fmt.Printf("Game over\n")
                fmt.Println("Player", game.Players[0].Name, "has won the game")
                break
            }
            PrintScoresOfRound(game)
			continue
		}
        ClearScreen()
        PrintScoresOfRound(game)
		round++
	}
}

func PlayRound(game Game, round int, scanner *bufio.Scanner) Player {
	playerCount := len(game.Players)
	scores := make([]float64, playerCount)
	targetScore := 0.0
	for i := 0; i < playerCount; i++ {
		fmt.Printf("\nPlayer %s : ( Enter your Choice ): \n", game.Players[i].Name)
		scores[i] = float64(ConstraintScanner(scanner, 0, 100))
		targetScore += scores[i]
	}

	targetScore = (targetScore * 0.8) / float64(playerCount)

	minValue :=  math.Abs(scores[0] - targetScore)
	winnerPlayer := game.Players[0]
	for i := 0; i < playerCount; i++ {
		if minValue > math.Abs(scores[i]-targetScore) {
			minValue = math.Abs(scores[i] - targetScore)
			winnerPlayer = game.Players[i]
		}
	}

	for i := 0; i < playerCount; i++ {
		if game.Players[i] != winnerPlayer {
			game.Players[i].Score++
		}
	}


	return winnerPlayer
}
