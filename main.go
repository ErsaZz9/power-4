package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const (
	Rows    = 6
	Cols    = 7
	Player1 = 1
	Player2 = 2
)

type GameState struct {
	Board         [Rows][Cols]int
	CurrentPlayer int
	Message       string
	IsOver        bool
	HasStarted    bool
	LastMoveRow   int
	LastMoveCol   int
}

var gameState GameState

func initGame() {
	gameState = GameState{
		CurrentPlayer: Player1,
		Message:       "Cliquez sur JOUER pour défier le Caca Maléfique ou casse toi !",
		IsOver:        false,
		HasStarted:    false,
		LastMoveRow:   -1,
		LastMoveCol:   -1,
	}
	rand.Seed(time.Now().UnixNano())
}

func dropToken(col int, player int) (row int, success bool) {
	if col < 0 || col >= Cols || gameState.IsOver {
		return -1, false
	}

	for r := Rows - 1; r >= 0; r-- {
		if gameState.Board[r][col] == 0 {
			gameState.Board[r][col] = player
			gameState.LastMoveRow = r
			gameState.LastMoveCol = col
			return r, true
		}
	}
	return -1, false
}

func checkWin(player int) bool {
	for r := 0; r < Rows; r++ {
		for c := 0; c <= Cols-4; c++ {
			if gameState.Board[r][c] == player &&
				gameState.Board[r][c+1] == player &&
				gameState.Board[r][c+2] == player &&
				gameState.Board[r][c+3] == player {
				return true
			}
		}
	}

	for c := 0; c < Cols; c++ {
		for r := 0; r <= Rows-4; r++ {
			if gameState.Board[r][c] == player &&
				gameState.Board[r+1][c] == player &&
				gameState.Board[r+2][c] == player &&
				gameState.Board[r+3][c] == player {
				return true
			}
		}
	}

	for r := 3; r < Rows; r++ {
		for c := 0; c <= Cols-4; c++ {
			if gameState.Board[r][c] == player &&
				gameState.Board[r-1][c+1] == player &&
				gameState.Board[r-2][c+2] == player &&
				gameState.Board[r-3][c+3] == player {
				return true
			}
		}
	}

	for r := 0; r <= Rows-4; r++ {
		for c := 0; c <= Cols-4; c++ {
			if gameState.Board[r][c] == player &&
				gameState.Board[r+1][c+1] == player &&
				gameState.Board[r+2][c+2] == player &&
				gameState.Board[r+3][c+3] == player {
				return true
			}
		}
	}

	return false
}

func checkDraw() bool {
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			if gameState.Board[r][c] == 0 {
				return false
			}
		}
	}
	return true
}

func robotMove() {
	availableCols := []int{}
	for c := 0; c < Cols; c++ {
		if gameState.Board[0][c] == 0 {
			availableCols = append(availableCols, c)
		}
	}

	if len(availableCols) > 0 {
		col := availableCols[rand.Intn(len(availableCols))]

		_, success := dropToken(col, Player2)
		if success {
			messages := []string{
				"Le caca maléfique réfléchit... puis joue.",
				"Attention, il prépare quelque chose !",
				"Le caca maléfique joue un coup au hasard... pour le moment.",
				"Trick or Treat! Le caca maléfique a joué.",
			}
			gameState.Message = messages[rand.Intn(len(messages))]
			return
		}
	}

	gameState.Message = "Le caca maléfique ne peut pas jouer, la grille est pleine !"
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	funcs := template.FuncMap{
		"makeRange": func(min, max int) []int {
			a := make([]int, max-min+1)
			for i := range a {
				a[i] = min + i
			}
			return a
		},
		"inc": func(i int) int {
			return i + 1
		},
	}

	tmpl, err := template.New("index.html").Funcs(funcs).ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Erreur de chargement du template: "+err.Error(), http.StatusInternalServerError)
		log.Println("Erreur de template:", err)
		return
	}

	if r.URL.Query().Get("start") == "true" && !gameState.HasStarted {
		initGame()
		gameState.HasStarted = true
		gameState.Message = "C'est ton tour bouffon(e) (Orange) !"
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.URL.Query().Get("restart") == "true" {
		initGame()
		gameState.HasStarted = true
		gameState.Message = "C'est ton tour bouffon(e) (Orange) !"
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if gameState.HasStarted && !gameState.IsOver && gameState.CurrentPlayer == Player2 {
		time.Sleep(500 * time.Millisecond)
        
		robotMove()

		if checkWin(Player2) {
			gameState.Message = "GAME OVER! Le Caca maléfique a gagné. looser..."
			gameState.IsOver = true
		} else if checkDraw() {
			gameState.Message = "MATCH NUL! Personne n'a gagné. bruh...t nul ou quoi"
			gameState.IsOver = true
		} else {
			gameState.CurrentPlayer = Player1
			gameState.Message = "C'est ton tour bouffon(e) (Orange) !"
		}
	}

	err = tmpl.Execute(w, gameState)
	if err != nil {
		http.Error(w, "Erreur d'exécution du template: "+err.Error(), http.StatusInternalServerError)
		log.Println("Erreur d'exécution:", err)
	}
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	if !gameState.HasStarted || gameState.IsOver {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Erreur de formulaire", http.StatusBadRequest)
		return
	}

	colStr := r.FormValue("column")
	col, err := strconv.Atoi(colStr)
	if err != nil {
		http.Error(w, "Colonne invalide", http.StatusBadRequest)
		return
	}

	_, success := dropToken(col, Player1)

	if success {
		if checkWin(Player1) {
			gameState.Message = "VICTOIRE! t'as vaincu le Caca Maléfique, au moins ça t'y arrive."
			gameState.IsOver = true
		} else if checkDraw() {
			gameState.Message = "MATCH NUL! Personne n'a gagné (même battre un bail aléatoire t'arrive pas looser)."
			gameState.IsOver = true
		} else {
			gameState.CurrentPlayer = Player2
		}
	} else {
		gameState.Message = "t'es bête ou quoi tu vois pas la colonne pleine ?"
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	initGame()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/play", playHandler)
    
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	port := ":8081"
	log.Printf("Serveur Trick4Treat démarré sur http://localhost%s", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("le serveur il veut pas wlh: %v", err)
	}
}
