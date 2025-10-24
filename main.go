// Package principal
package main

// Importation des librairies nécessaires
import (
	// Pour les fonctions de formatage (comme fmt.Println)
	"html/template" // Pour gérer les templates HTML
	"log"           // Pour l'affichage des logs et des erreurs
	"math/rand"     // Pour la fonction aléatoire de l'IA (robot)
	"net/http"      // Pour créer le serveur HTTP
	"strconv"       // Pour convertir les chaînes en nombres (comme la colonne choisie)
	"time"          // Pour initialiser la graine aléatoire de rand
)

// Constantes du jeu
const (
	Rows    = 6 // Nombre de lignes (hauteur du plateau)
	Cols    = 7 // Nombre de colonnes (largeur du plateau)
	Player1 = 1 // Représentation du joueur humain
	Player2 = 2 // Représentation du robot/IA
)

// Structure pour stocker l'état global du jeu
type GameState struct {
	Board         [Rows][Cols]int // Le plateau de jeu (0:vide, 1:J1, 2:J2)
	CurrentPlayer int             // Le joueur dont c'est le tour (1 ou 2)
	Message       string          // Message affiché à l'utilisateur (victoire, tour, etc.)
	IsOver        bool            // Indique si la partie est terminée
	HasStarted    bool            // Indique si le joueur a cliqué sur 'JOUER'
	LastMoveType  string          // Type du dernier coup du robot ("normal" ou "special")
}

// État du jeu initialisé globalement
var gameState GameState

// Initialise le plateau de jeu à vide et le premier joueur (humain)
func initGame() {
	gameState = GameState{
		CurrentPlayer: Player1,
		Message:       "Cliquez sur JOUER pour défier le Robot Hanté !",
		IsOver:        false,
		HasStarted:    false,
		LastMoveType:  "none",
	}
	// Initialiser la graine aléatoire pour l'IA
	rand.Seed(time.Now().UnixNano())
}

// Fonction pour déterminer la ligne où le jeton va tomber dans une colonne donnée
func dropToken(col int, player int) (row int, success bool) {
	if col < 0 || col >= Cols || gameState.IsOver {
		return -1, false // Colonne invalide ou jeu terminé
	}

	// Parcourt les lignes de bas en haut
	for r := Rows - 1; r >= 0; r-- {
		if gameState.Board[r][col] == 0 {
			// Trouve la première case vide
			gameState.Board[r][col] = player
			return r, true
		}
	}
	return -1, false // Colonne pleine
}

// Fonction pour vérifier si le joueur donné a gagné
func checkWin(player int) bool {
	// Vérification horizontale (4 jetons alignés)
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

	// Vérification verticale
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

	// Vérification diagonale (bas-gauche vers haut-droite)
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

	// Vérification diagonale (haut-gauche vers bas-droite)
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

// Fonction pour vérifier si le plateau est plein (match nul)
func checkDraw() bool {
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			if gameState.Board[r][c] == 0 {
				return false // Au moins une case vide
			}
		}
	}
	return true // Plateau plein
}

// --- LOGIQUE D'INTELLIGENCE ARTIFICIELLE (IA) ---

// Fonction qui exécute le coup du robot (Player2)
func robotMove() {
	gameState.LastMoveType = "normal"

	// Option 1: Coup spécial (Vol de place) - 1 chance sur 10
	if rand.Intn(10) == 0 {
		// Tente de trouver un jeton du joueur 1 à voler
		for r := Rows - 1; r >= 0; r-- {
			for c := 0; c < Cols; c++ {
				if gameState.Board[r][c] == Player1 {
					// Voler le jeton du Joueur 1 (le remplacer par un jeton 2)
					gameState.Board[r][c] = Player2
					gameState.Message = "Le Robot a utilisé Trick! Il a volé votre jeton en colonne " + strconv.Itoa(c+1) + " !"
					gameState.LastMoveType = "special"
					return // Le coup spécial est exécuté, fin du tour
				}
			}
		}
	}

	// Option 2: Coup normal (Aléatoire dans une colonne non pleine)
	// Trouve les colonnes jouables
	availableCols := []int{}
	for c := 0; c < Cols; c++ {
		// Si la colonne n'est pas pleine (vérifie la ligne du haut)
		if gameState.Board[0][c] == 0 {
			availableCols = append(availableCols, c)
		}
	}

	if len(availableCols) > 0 {
		// Choisit une colonne aléatoire parmi les jouables
		col := availableCols[rand.Intn(len(availableCols))]

		// Lâche le jeton (la ligne est déterminée dans dropToken)
		_, success := dropToken(col, Player2)
		if success {
			if gameState.LastMoveType != "special" {
				// Message par défaut si ce n'était pas un coup spécial
				messages := []string{
					"Le Robot réfléchit... puis joue.",
					"Attention, il prépare quelque chose !",
					"Le Robot joue un coup au hasard... pour le moment.",
					"Trick or Treat! Le Robot a joué.",
				}
				gameState.Message = messages[rand.Intn(len(messages))]
			}
			return
		}
	}

	// Si toutes les colonnes sont pleines (cas impossible si on a déjà vérifié pour le Draw)
	gameState.Message = "Le Robot ne peut pas jouer, la grille est pleine !"
}

// --- HANDLERS HTTP ---

// Handler pour la route principale (GET /)
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Fonction utilitaire pour le template (range de 0 à 6)
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

	// Charger le template HTML
	tmpl, err := template.New("index.html").Funcs(funcs).ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Erreur de chargement du template: "+err.Error(), http.StatusInternalServerError)
		log.Println("Erreur de template:", err)
		return
	}

	// Gestion du démarrage du jeu (clic sur le bouton JOUER)
	if r.URL.Query().Get("start") == "true" && !gameState.HasStarted {
		initGame() // Réinitialiser le jeu (même si pas fini)
		gameState.HasStarted = true
		gameState.Message = "C'est votre tour (Orange) !"
		http.Redirect(w, r, "/", http.StatusSeeOther) // Redirige pour nettoyer l'URL
		return
	}

	// Gestion du redémarrage du jeu
	if r.URL.Query().Get("restart") == "true" {
		initGame()
		gameState.HasStarted = true
		gameState.Message = "C'est votre tour (Orange) !"
		http.Redirect(w, r, "/", http.StatusSeeOther) // Redirige pour nettoyer l'URL
		return
	}

	// Si le jeu a commencé et que c'est le tour du robot, jouer le coup du robot
	if gameState.HasStarted && !gameState.IsOver && gameState.CurrentPlayer == Player2 {
		robotMove() // Le robot joue

		// Vérification après le coup du robot
		if checkWin(Player2) {
			gameState.Message = "GAME OVER! Le Robot Hanté a gagné. Trick!"
			gameState.IsOver = true
		} else if checkDraw() {
			gameState.Message = "MATCH NUL! Personne n'a gagné."
			gameState.IsOver = true
		} else {
			// Le tour revient au joueur humain
			gameState.CurrentPlayer = Player1
			gameState.Message = "C'est votre tour (Orange) !"
		}
	}

	// Rendu de la page avec l'état actuel du jeu
	err = tmpl.Execute(w, gameState)
	if err != nil {
		http.Error(w, "Erreur d'exécution du template: "+err.Error(), http.StatusInternalServerError)
		log.Println("Erreur d'exécution:", err)
	}
}

// Handler pour la route de jeu (POST /play)
func playHandler(w http.ResponseWriter, r *http.Request) {
	// Assurez-vous que la méthode est bien POST
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Si le jeu n'a pas commencé ou est terminé, on ne fait rien
	if !gameState.HasStarted || gameState.IsOver {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// 1. Récupérer la colonne soumise par le joueur
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

	// 2. Tenter de placer le jeton
	_, success := dropToken(col, Player1)

	if success {
		// 3. Le coup est valide

		// Vérification après le coup du joueur humain
		if checkWin(Player1) {
			gameState.Message = "VICTOIRE! Vous avez vaincu le Robot Hanté! Treat!"
			gameState.IsOver = true
		} else if checkDraw() {
			gameState.Message = "MATCH NUL! Personne n'a gagné."
			gameState.IsOver = true
		} else {
			// Tour du Robot
			gameState.CurrentPlayer = Player2
			// Le message sera mis à jour par le homeHandler lorsque le robot joue
		}
	} else {
		// Le coup n'est pas valide (colonne pleine)
		gameState.Message = "Colonne pleine, veuillez choisir une autre colonne (Orange)."
	}

	// Rediriger vers la page principale (GET /) pour rafraîchir le plateau
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Fonction principale : démarrage du serveur
func main() {
	// 1. Initialiser l'état du jeu
	initGame()

	// 2. Définir les routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/play", playHandler)

	// 3. Démarrer le serveur HTTP
	// Nous utilisons le port 8081 pour éviter le conflit avec le port 8080 qui était bloqué
	port := ":8081"
	log.Printf("Serveur Trick4Treat démarré sur http://localhost%s", port)

	// Écoute des connexions et gestion des erreurs
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Impossible de démarrer le serveur: %v", err)
	}
}
