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

var DB sql.DB = openDB()

func openDB() sql.DB{
  DB, err := sql.Open("sqlite3", "./database.db")
  if err != nil {
   fmt.Println(err.Error())
  }
  return *DB
}

func fillDB(ml []*discordgo.Member){
  for i := 0; i<len(ml); i++ {
    ID := ml[i].User.ID
    UN := ml[i].User.Username
    Disc := ml[i].User.Discriminator
    sqlStmt, err := DB.Prepare("INSERT INTO users (id, name, discrim, exp, vexp) values (?, ?, ?, '0', '0');")
    if err != nil {
     fmt.Println(err.Error())
    }
    sqlStmt.Exec(ID, UN, Disc)
	}
}

func addExp(m *discordgo.MessageCreate){
  u := new(user)
  err := DB.QueryRow("SELECT * FROM users WHERE id = ?;", m.Author.ID).Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp)
  if err != nil {
    fmt.Println(err.Error())
  }
  sqlStmt, err := DB.Prepare("UPDATE users SET exp = ? WHERE id = ?;")
  if err != nil {
    fmt.Println(err.Error())
  }
  u.exp = u.exp + 10
  sqlStmt.Exec(u.exp,m.Author.ID)
}

func printLeaderboard(s *discordgo.Session, m *discordgo.MessageCreate){
  u := new(user)
  message := "```\nTop 10 Users:\n"
  rows, _ := DB.Query("SELECT * FROM users ORDER BY exp DESC LIMIT 10;")
  defer rows.Close()
  for rows.Next(){
    rows.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp)
    message = message + u.name + ": " + strconv.Itoa(u.exp) + "\n"
  }
  message = message + "\n```"
  _, _ = s.ChannelMessageSend(m.ChannelID, message)
}

func findExp(m *discordgo.MessageCreate) (int, int){
  u := new(user)
  err := DB.QueryRow("SELECT * FROM users WHERE id = ?;", m.Author.ID).Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp)
  if err != nil {
    fmt.Println(err.Error())
  }
  return u.exp, u.vexp
}

func findPos(m *discordgo.MessageCreate, exp int) int{
  var ranking int
  err := DB.QueryRow("SELECT COUNT (*) FROM users WHERE exp >= ?;", exp).Scan(&ranking)
  if err != nil {
    fmt.Println(err.Error())
  }
  return ranking
}

func addTimeToDB(time int64, m *discordgo.VoiceStateUpdate){
  u := new(user)
  err := DB.QueryRow("SELECT * FROM users WHERE id = ?;", m.UserID).Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp)
  if err != nil {
    fmt.Println(err.Error())
  }
  sqlStmt, err := DB.Prepare("UPDATE users SET vexp = ? WHERE id = ?;")
  if err != nil {
    fmt.Println(err.Error())
  }
  fmt.Println(u.vexp)
  time = time + int64(u.vexp)
  sqlStmt.Exec(time,m.UserID)
}
