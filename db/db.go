package db
import(
    "database/sql"
    _"github.com/mattn/go-sqlite3"
)

// Main represents the path of main database
const Main = "./db/main.db"

// Connect is used to connect to a database by specifying path
func Connect(path string) (*sql.DB, error){
    return sql.Open("sqlite3", path)
}
