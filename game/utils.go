package game

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

func ClearScreen() {

	var cmd *exec.Cmd
	cmd = exec.Command("clear")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}

	if cmd.Err != nil {
		return
	}

	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return
	}
}

func CheckPlayerElimination(game Game) (bool, int) {
    for idx := range game.Players {
        if game.Players[idx].Score >= game.EliminationScore {
            return true, idx
        }
    }
    return false, 0
}

func ConstraintScanner(scanner *bufio.Scanner, min, max int) int {
    var output int
    for {
        scanner.Scan()
        counter, err := strconv.Atoi(scanner.Text())

        if err != nil {
            fmt.Print("Invalid input\n")
            continue
        }
        if counter < min || counter > max {
            fmt.Printf("Invalid input %d\n\nSelect a number between %d and %d\n", counter, min, max)
            continue
        }
        output = counter
        break
    }

    return output
}

func PrintScoresOfRound(game Game){
    maxNameLength := 4
    for _, player := range game.Players {
        maxNameLength = max(maxNameLength, len(player.Name))
    }
    fmt.Print("===== Round Results =====\n\n")
    fmt.Print(" Name")
    for i := 0; i < maxNameLength - 3; i++ {
        fmt.Print(" ")
    }
    fmt.Print("| Score \n")
    for i := 0; i < maxNameLength + 2; i++ {
        fmt.Print("-")
    }
    fmt.Print("|")
    for i := 0; i < 5; i++ {
        fmt.Print("-")
    }
    fmt.Print("\n")
    for _, player := range game.Players {
        fmt.Printf(" %s", player.Name)
        for i := 0; i < maxNameLength - len(player.Name) + 1; i++ {
            fmt.Print(" ")
        }
        fmt.Printf("| %d \n", player.Score)
    }
    fmt.Print("\n")
}
