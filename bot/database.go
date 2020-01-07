package bot

import (
   "fmt"
   "database/sql"
   "github.com/bwmarrin/discordgo"
   _ "github.com/mattn/go-sqlite3"
   "strconv"
)

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
  db, err := sql.Open("sqlite3", "./database.db")
  if err != nil {
    fmt.Println(err.Error())
  }
  rows, err := db.Query("SELECT * FROM users WHERE id = ?;", m.Author.ID)
  if err != nil {
    fmt.Println(err.Error())
  }
  var id string
  var name string
  var discrim string
  var vexp int
  var exp int
  defer rows.Close()
  for rows.Next(){
    rows.Scan(&id, &name, &discrim, &exp, &vexp)
    }
  sqlStmt, err := db.Prepare("UPDATE users SET exp = ? WHERE id = ?;")
  if err != nil {
    fmt.Println(err.Error())
  }
  exp = exp+10
  sqlStmt.Exec(exp,m.Author.ID)
  db.Close()
}

func printLeaderboard(s *discordgo.Session, m *discordgo.MessageCreate){
  message := "```\nTop 10 Users:\n"
  db, _ := sql.Open("sqlite3", "./database.db")
  rows, _ := db.Query("SELECT * FROM users ORDER BY exp DESC LIMIT 10;")
  var id string
  var name string
  var discrim string
  var vexp int
  var exp int
  defer rows.Close()
  for rows.Next(){
    rows.Scan(&id, &name, &discrim, &exp, &vexp)
    message = message + name + ": " + strconv.Itoa(exp) + "\n"
  }
  message = message + "\n```"
  _, _ = s.ChannelMessageSend(m.ChannelID, message)
}
