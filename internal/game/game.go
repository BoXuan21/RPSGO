package game

type Game struct {
	PlayerChoice   string `json:"player_choice"`
	ComputerChoice string `json:"computer_choice"`
	Result         string `json:"result"`
}

type Stats struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
}

var Choices = []string{"rock", "paper", "scissors"}

func DetermineWinner(player, computer string) string {
	if player == computer {
		return "draw"
	}

	if (player == "rock" && computer == "scissors") ||
		(player == "paper" && computer == "rock") ||
		(player == "scissors" && computer == "paper") {
		return "win"
	}

	return "lose"
}

func ValidateChoice(choice string) bool {
	for _, validChoice := range Choices {
		if choice == validChoice {
			return true
		}
	}
	return false
}
