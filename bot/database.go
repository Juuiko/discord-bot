package bot

import (
   "fmt"
   "database/sql"
   "github.com/bwmarrin/discordgo"
   _ "github.com/mattn/go-sqlite3"
   "strconv"
)

type user struct{
  id string
  name string
  discrim string
  vexp int
  exp int
}

func fillDB(ml []*discordgo.Member){
  db, err := sql.Open("sqlite3", "./database.db")
  if err != nil {
   fmt.Println(err.Error())
  }
  for i := 0; i<len(ml); i++ {
    ID := ml[i].User.ID
    UN := ml[i].User.Username
    Disc := ml[i].User.Discriminator
    sqlStmt, err := db.Prepare("INSERT INTO users (id, name, discrim, exp, vexp) values (?, ?, ?, '0', '0');")
    if err != nil {
     fmt.Println(err.Error())
    }
    sqlStmt.Exec(ID, UN, Disc)
	}
  db.Close()
}

func addExp(m *discordgo.MessageCreate){
  u := new(user)
  db, err := sql.Open("sqlite3", "./database.db")
  if err != nil {
    fmt.Println(err.Error())
  }
  rows, err := db.Query("SELECT * FROM users WHERE id = ?;", m.Author.ID)
  if err != nil {
    fmt.Println(err.Error())
  }
  defer rows.Close()
  for rows.Next(){
    rows.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp)
    }
  sqlStmt, err := db.Prepare("UPDATE users SET exp = ? WHERE id = ?;")
  if err != nil {
    fmt.Println(err.Error())
  }
  u.exp = u.exp + 10
  sqlStmt.Exec(u.exp,m.Author.ID)
  db.Close()
}

func printLeaderboard(s *discordgo.Session, m *discordgo.MessageCreate){
  u := new(user)
  message := "```\nTop 10 Users:\n"
  db, _ := sql.Open("sqlite3", "./database.db")
  rows, _ := db.Query("SELECT * FROM users ORDER BY exp DESC LIMIT 10;")
  defer rows.Close()
  for rows.Next(){
    rows.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp)
    message = message + u.name + ": " + strconv.Itoa(u.exp) + "\n"
  }
  message = message + "\n```"
  _, _ = s.ChannelMessageSend(m.ChannelID, message)
  db.Close()
}

func findExp(m *discordgo.MessageCreate) int{
  u := new(user)
  db, err := sql.Open("sqlite3", "./database.db")
  if err != nil {
    fmt.Println(err.Error())
  }
  err = db.QueryRow("SELECT * FROM users WHERE id = ?;", m.Author.ID).Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp)
  if err != nil {
    fmt.Println(err.Error())
  }
  db.Close()
  return u.exp
}

func findPos(m *discordgo.MessageCreate, exp int) int{
  var ranking int
  db, err := sql.Open("sqlite3", "./database.db")
  if err != nil {
    fmt.Println(err.Error())
  }
  err = db.QueryRow("SELECT COUNT (*) FROM users WHERE exp >= ?;", exp).Scan(&ranking)
  if err != nil {
    fmt.Println(err.Error())
    fmt.Println("reeeee")
  }
  db.Close()
  return ranking
}
