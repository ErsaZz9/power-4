package main

import (
	"fmt"
	"html/template"
	"net/http"
)

// Structure de données envoyée au HTML
type PageData struct {
	Titre string
}

// Fonction principale
func main() {
	// Chargement du template HTML (index.html)
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	// Route principale "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{Titre: "👻 Power4 Halloween 🎃"} // Données à envoyer
		tmpl.Execute(w, data)                           // Affiche le HTML
	})

	// Route pour les fichiers statiques (CSS, vidéo, images…)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Lancer le serveur sur localhost:8080
	fmt.Println("Serveur lancé sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
