package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	tmplPath := filepath.Join("templates", "index.html")
	tmpl := template.Must(template.ParseFiles(tmplPath))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, "Erreur serveur", http.StatusInternalServerError)
			log.Println("template execute:", err)
		}
	})

	// Servir les fichiers statiques (CSS, vidéo...)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	addr := ":8080"
	log.Printf("Serveur démarré sur http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}