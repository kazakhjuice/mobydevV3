package auth

type FilmDetails struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	ProjectType string   `json:"project_type"`
	Year        string   `json:"year"`
	Duration    int      `json:"duration"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	Director    string   `json:"director"`
	Producer    string   `json:"producer"`
}
