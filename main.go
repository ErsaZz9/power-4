package main

import (
	"html/template"
	"log"
	"net/http"
)

// PageData si besoin d'envoyer des champs (tu peux étendre)
type PageData struct {
	Title string
}

func main() {
	// Parse le template index.html (erreur panic si absent)
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	// Route racine : exécute le template
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{Title: "Trick4Treat"}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Erreur template", http.StatusInternalServerError)
			log.Println("Template execute error:", err)
		}
	})

	// Servir les fichiers statiques (css, videos, images, sounds)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Serveur lancé sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
