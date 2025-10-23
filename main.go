// Fichier : main.go
// Description : Serveur HTTP Go pour le jeu Puissance 4, avec gestion de l'état
// en mémoire (SSR - Server-Side Rendering) et un thème Halloween.
package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Constantes du jeu
const (
	Rows    = 6
	Cols    = 7
	Player1 = 1 // Violet
	Player2 = 2 // Orange
	Empty   = 0
)

// --- Structures de données ---

// Game représente l'état actuel du Puissance 4
type Game struct {
	Board         [Rows][Cols]int // Le plateau de jeu (0: vide, 1: J1, 2: J2)
	CurrentPlayer int             // Le joueur dont c'est le tour (1 ou 2)
	Winner        int             // 0: en cours, 1: J1 gagne, 2: J2 gagne, 3: Match Nul
	Message       string          // Message affiché à l'utilisateur
}

// Data pour le template HTML
type TemplateData struct {
	Game
	// Les fonctions utilitaires sont ajoutées via template.FuncMap
}

// --- État Global du Jeu ---

// Utilisation d'une variable globale pour l'état du jeu (jouer sur le "même ordinateur")
var currentGame Game
var mu sync.Mutex // Mutex pour protéger l'accès concurrent à l'état du jeu

// --- Helpers pour Go Templates ---

// makeSlice crée un tableau d'entiers [0, 1, 2, ..., n-1]
// Utilisé dans le HTML avec {{range $col := makeSlice 7}}
func makeSlice(n int) []int {
	s := make([]int, n)
	for i := 0; i < n; i++ {
		s[i] = i
	}
	return s
}

// addOne ajoute +1 à un nombre (utile pour afficher les numéros de colonnes 1,2,3...)
// Utilisé dans le HTML avec {{$col | addOne}}
func addOne(i int) int {
	return i + 1
}

// --- Logique du Jeu ---

// initGame initialise une nouvelle partie
func initGame() {
	mu.Lock()
	defer mu.Unlock()

	currentGame = Game{
		Board:         [Rows][Cols]int{}, // Plateau vide
		CurrentPlayer: Player1,
		Winner:        0,
		Message:       "🎃 Bienvenue sur le plateau maudit ! C’est au tour du joueur Violet 👻",
	}
	rand.Seed(time.Now().UnixNano())
	log.Println("Nouvelle partie initialisée.")
}

// dropPiece place un jeton dans la colonne spécifiée.
// Renvoie true si le coup est valide, false sinon.
func (g *Game) dropPiece(col int) bool {
	if g.Winner != 0 {
		return false // Le jeu est terminé
	}
	if col < 0 || col >= Cols {
		return false // Colonne invalide
	}

	// Chercher la première ligne vide en partant du bas
	for row := Rows - 1; row >= 0; row-- {
		if g.Board[row][col] == Empty {
			g.Board[row][col] = g.CurrentPlayer
			return true
		}
	}
	// La colonne est pleine
	return false
}

// checkWin vérifie si le dernier coup a entraîné une victoire ou un match nul.
func (g *Game) checkWin() {
	player := g.CurrentPlayer

	// Fonction utilitaire pour vérifier 4 jetons alignés
	checkLine := func(r, c, dr, dc int) bool {
		count := 0
		for i := 0; i < 4; i++ {
			row, col := r+i*dr, c+i*dc
			if row >= 0 && row < Rows && col >= 0 && col < Cols && g.Board[row][col] == player {
				count++
			} else {
				return false // Arrêter si le joueur n'est pas le bon ou si hors limites
			}
		}
		return count == 4
	}

	// 1. Vérification des Alignements (Horizontal, Vertical, Diagonales)
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			// Vérifications à partir de chaque cellule (pour être la première d'un alignement)
			if g.Board[r][c] == player {
				// Horizontal (vers la droite)
				if c <= Cols-4 && checkLine(r, c, 0, 1) {
					g.Winner = player
					return
				}
				// Vertical (vers le bas, inutile car jetons tombent, mais pour la symétrie)
				if r <= Rows-4 && checkLine(r, c, 1, 0) {
					g.Winner = player
					return
				}
				// Diagonale descendante (vers bas-droite)
				if r <= Rows-4 && c <= Cols-4 && checkLine(r, c, 1, 1) {
					g.Winner = player
					return
				}
				// Diagonale ascendante (vers haut-droite)
				if r >= Rows-4 && c <= Cols-4 && checkLine(r, c, -1, 1) {
					g.Winner = player
					return
				}
			}
		}
	}

	// 2. Vérification du Match Nul
	isBoardFull := true
	for c := 0; c < Cols; c++ {
		if g.Board[0][c] == Empty {
			isBoardFull = false // Il reste au moins une case vide
			break
		}
	}
	if isBoardFull && g.Winner == 0 {
		g.Winner = 3 // Code 3 pour le Match Nul
	}
}

// getThemedMessage génère un message d'ambiance Halloween
func getThemedMessage(player int, winner int) string {
	if winner == Player1 {
		return "VICTOIRE ! Le fantôme Violet a ensorcelé le plateau ! 💜"
	}
	if winner == Player2 {
		return "VICTOIRE ! La citrouille Orange a terrifié son adversaire ! 🧡"
	}
	if winner == 3 {
		return "MATCH NUL. Les esprits ne se départagent pas. Réinitialisez pour un nouveau combat."
	}

	// Messages aléatoires pour le tour suivant
	playerColor := "Violet"
	if player == Player2 {
		playerColor = "Orange"
	}

	messages := []string{
		fmt.Sprintf("Au tour du joueur %s. Le cimetière vous attend...", playerColor),
		fmt.Sprintf("Joueur %s, le chaudron fume... faites votre coup !", playerColor),
		fmt.Sprintf("Le robot vous observe, joueur %s. Ne tremblez pas.", playerColor),
		fmt.Sprintf("Un coup de poignard, joueur %s ! (ou juste une pièce, c'est comme vous voulez).", playerColor),
	}
	return messages[rand.Intn(len(messages))]
}

// --- Handlers HTTP ---

// handleRoot sert la page HTML principale du jeu.
func handleRoot(w http.ResponseWriter, r *http.Request) {
	// S'assurer que le jeu est initialisé au premier accès
	mu.Lock()
	defer mu.Unlock()

	if currentGame.CurrentPlayer == 0 {
		initGame()
	}

	// Le routeur gère les fichiers statiques (CSS, images, etc.) sur /static/
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Création du map de fonctions pour le template
	funcMap := template.FuncMap{
		"makeSlice": makeSlice,
		"addOne":    addOne,
	}

	// Chargement et parsing du template (effectué à chaque requête pour simplifier le développement)
	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Erreur de chargement du template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := TemplateData{Game: currentGame}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

// handlePlay reçoit le coup du joueur et met à jour l'état du jeu.
func handlePlay(w http.ResponseWriter, r *http.Request) {
	// S'assurer que la méthode est bien POST, comme spécifié par la consigne
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// ParseForm doit être appelé avant d'accéder aux valeurs du formulaire
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erreur de parsing du formulaire", http.StatusBadRequest)
		return
	}

	// Récupérer la colonne soumise par le formulaire
	colStr := r.FormValue("column")
	col, err := strconv.Atoi(colStr)
	if err != nil {
		// Ce cas arrive si la valeur du formulaire n'est pas un nombre
		http.Error(w, "Colonne invalide", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// 1. Tenter de jouer le coup
	if currentGame.Winner == 0 && currentGame.dropPiece(col) {
		// 2. Vérifier l'état du jeu (Victoire ou Nul)
		currentGame.checkWin()

		// 3. Mettre à jour le tour ou le message final
		if currentGame.Winner == 0 {
			// Changer de joueur
			if currentGame.CurrentPlayer == Player1 {
				currentGame.CurrentPlayer = Player2
			} else {
				currentGame.CurrentPlayer = Player1
			}
			// Définir un nouveau message thématique
			currentGame.Message = getThemedMessage(currentGame.CurrentPlayer, 0)
		} else {
			// Définir le message de victoire/nul
			currentGame.Message = getThemedMessage(0, currentGame.Winner)
		}
	} else if currentGame.Winner != 0 {
		currentGame.Message = getThemedMessage(0, currentGame.Winner)
	} else {
		// Le coup était invalide (colonne pleine, etc.)
		currentGame.Message = "Alerte fantôme ! Cette colonne est déjà pleine. Choisissez une autre."
	}

	// Rediriger vers la page principale pour afficher le nouvel état du plateau
	// Utilisation de StatusSeeOther (303) est une bonne pratique après un POST
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleReset réinitialise la partie.
func handleReset(w http.ResponseWriter, r *http.Request) {
	initGame()
	// Rediriger vers la page principale pour afficher le plateau initial
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// --- Fonction Principale ---

func main() {
	// Initialiser la première partie au démarrage du serveur
	initGame()

	// Handler pour les fichiers statiques (CSS, images, sons, vidéos)
	// http.StripPrefix retire le segment "/static/" de l'URL avant de chercher le fichier dans le dossier "static"
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Configuration des routes HTTP
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/play", handlePlay)
	http.HandleFunc("/reset", handleReset)

	// Démarrage du serveur
	fmt.Println("Serveur Puissance 4 Halloween lancé sur http://localhost:8080 🎃")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
