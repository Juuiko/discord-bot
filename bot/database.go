package bot

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
)

type user struct {
	id      string
	name    string
	discrim string
	vexp    int
	exp     int
	wvexp    int
	wexp     int
	mvexp    int
	mexp     int
}

var DB sql.DB = openDB()

func openDB() sql.DB {
	DB, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		fmt.Println(err.Error())
	}
	return *DB
}

func fillDB(ml []*discordgo.Member) {
	for i := 0; i < len(ml); i++ {
		sqlStmt, err := DB.Prepare("INSERT INTO users (id, name, discrim, exp, vexp, wexp, wvexp, mexp, mvexp) values (?, ?, ?, '0', '0', '0', '0', '0', '0');")
		if err != nil {
			fmt.Println(err.Error())
		}
		sqlStmt.Exec(ml[i].User.ID, ml[i].User.Username, ml[i].User.Discriminator)
	}
}

func addNewUser(u *discordgo.User)  {
	sqlStmt, err := DB.Prepare("INSERT INTO users (id, name, discrim, exp, vexp, wexp, wvexp, mexp, mvexp) values (?, ?, ?, '0', '0', '0', '0', '0', '0');")
	if err != nil {
		fmt.Println(err.Error())
	}
	sqlStmt.Exec(u.ID, u.Username, u.Discriminator)
}

func addExp(m *discordgo.MessageCreate) {
	u := new(user)
	err := DB.QueryRow("SELECT * FROM users WHERE id = ?;", m.Author.ID).Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp)
	if err != nil {
		fmt.Println(err.Error())
	}
	sqlStmt, err := DB.Prepare("UPDATE users SET exp = ?, wexp = ?, mexp = ? WHERE id = ?;")
	if err != nil {
		fmt.Println(err.Error())
	}
	u.exp = u.exp + 10
	u.wexp = u.wexp + 10
	u.mexp = u.mexp + 10
	sqlStmt.Exec(u.exp, u.wexp, u.mexp, m.Author.ID)
}

func printLeaderboard(s *discordgo.Session, m *discordgo.MessageCreate) {
	u := new(user)
	message := "```\nTop 10 Users:\n"
	rows, _ := DB.Query("SELECT * FROM users ORDER BY exp DESC LIMIT 10;")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp)
		message = message + u.name + ": " + strconv.Itoa(u.exp) + "\n"
	}
	message = message + "\n```"
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
}

func printVCLeaderboard(s *discordgo.Session, m *discordgo.MessageCreate) {
	u := new(user)
	message := "```\nTop 10 Voice Chat Users:\n"
	rows, _ := DB.Query("SELECT * FROM users ORDER BY vexp DESC LIMIT 10;")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp)
		message = message + u.name + ": " + secsToHours(u.vexp) + "\n"
	}
	message = message + "\n```"
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
}

func findExp(m *discordgo.MessageCreate) (int, int) {
	u := new(user)
	err := DB.QueryRow("SELECT * FROM users WHERE id = ?;", m.Author.ID).Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp)
	if err != nil {
		fmt.Println(err.Error())
	}
	return u.exp, u.vexp
}

func findPos(m *discordgo.MessageCreate, exp int) int {
	var ranking int
	err := DB.QueryRow("SELECT COUNT (*) FROM users WHERE exp >= ?;", exp).Scan(&ranking)
	if err != nil {
		fmt.Println(err.Error())
	}
	return ranking
}

func findVCPos(m *discordgo.MessageCreate, vexp int) int {
	var ranking int
	err := DB.QueryRow("SELECT COUNT (*) FROM users WHERE vexp >= ?;", vexp).Scan(&ranking)
	if err != nil {
		fmt.Println(err.Error())
	}
	return ranking
}

func addTimeToDB(time int64, m *discordgo.VoiceStateUpdate) {
	u := new(user)
	err := DB.QueryRow("SELECT * FROM users WHERE id = ?;", m.UserID).Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp)
	if err != nil {
		fmt.Println(err.Error())
	}
	sqlStmt, err := DB.Prepare("UPDATE users SET vexp = ?, wvexp = ?, mvexp = ? WHERE id = ?;")
	if err != nil {
		fmt.Println(err.Error())
	}
	dailyTime := time + int64(u.vexp)
	weeklyTime := time + int64(u.wvexp)
	monthlyTime := time + int64(u.mvexp)
	sqlStmt.Exec(dailyTime, weeklyTime, monthlyTime, m.UserID)
}
