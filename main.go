package main

import (
    "beauty/game"
)

func main() {
    gameInstance, scanner := game.InitGame()
    game.PlayGame(gameInstance, scanner)
}
