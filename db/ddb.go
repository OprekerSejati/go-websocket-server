package db

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql" // import ini dulu, JANGAN LUPA
	"golang.org/x/crypto/bcrypt"
)


var DB *sql.DB  
  
func InitDB(dataSourceName string) {  

	var err error  
	DB, err = sql.Open("mysql", dataSourceName)  
	if err != nil {  
		log.Fatal("Error connecting to the database:", err)  
	}  
  
	// Test the connection  
	if err = DB.Ping(); err != nil {  
		log.Fatal("Error pinging the database:", err)  
	}    
}  
  

func VerifyUser(username, password string) (bool, error) {  
	var storedHash string  
  
	err := DB.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&storedHash)  
	if err != nil {  
		if err == sql.ErrNoRows {  
			// Username not found
			return false, nil  
		}  
		return false, err  
	}  
  
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))  
	if err != nil {  
		// Password error
		return false, nil  
	}  
  
	// Username dan password cocok  
	return true, nil  
}  


func RegisterUser(username, password string) error {  
	// Hash password  
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)  
	if err != nil {  
		return err  
	}  
  
	_, err = DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, hashedPassword)  
	return err  
} 

func SaveMessage(username, room, message, receiver string) error {
	log.Println("SaveMessage username :", username)
	_, err := DB.Exec("INSERT INTO messages (messages,room,sender,receiver) VALUES (?, ?, ?, ?)", message,room, username, receiver)
 
	return err  
}  
