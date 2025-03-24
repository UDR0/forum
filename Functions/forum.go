package forum

import (
	"database/sql"
	"fmt"
)

func SayHello() {
	fmt.Println("Hello from forum.go!")
}

// ////////TO DO LIST////////////////
// get the infos from the database (ex => GetImageURLFromDB())
// write into the database (ex => WriteImageIntoDB())
// make sure all the tables in the database are right (don't want to)
// write all the url to get to the images

func GetImageURLFromDB() string {
	// Open the SQLite database located at /forum.db
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return ""
	}
	defer db.Close()

	// Query the image_url for id_region = 8
	var imageURL string
	err = db.QueryRow("SELECT image_url FROM regions WHERE id_region = 8").Scan(&imageURL)
	if err != nil {
		fmt.Printf("Failed to query image_url: %v\n", err)
		return ""
	}

	return imageURL
}

func WriteImageIntoDB() {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	// Write (INSERT) data into the database
	query := `SELECT * FROM regions; UPDATE regions SET image_url = 'peoedj' WHERE id_region = 1;`
	_, err = db.Exec(query)
	if err != nil {
		fmt.Println("Error executing query:", err)
		return
	}

	fmt.Println("Data successfully inserted into the database!")
}

/*    HOW TO SHOW THE IMAGE ON THE REGION.TMPL FILE BUT I NEED TO CHANGE THE NAME IN THE DATABASE
func forumHandler(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("templates", "region.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		ImageURL string
	}{
		ImageURL: getImageURLFromDB(),
	}

	tmpl.Execute(w, data)
}
*/
