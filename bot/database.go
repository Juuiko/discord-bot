package bot

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3" //SQLite 3 import
)

type user struct {
	id      string
	name    string
	discrim string
	vexp    int
	exp     int
	wvexp   int
	wexp    int
	mvexp   int
	mexp    int
	aml     float32
	amlc    int
}

//DB is open SQLite database
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
		sqlStmt, err := DB.Prepare("INSERT INTO users (id, name, discrim, exp, vexp, wexp, wvexp, mexp, mvexp, aml, amlc) values (?, ?, ?, '0', '0', '0', '0', '0', '0', '0', '0');")
		if err != nil {
			fmt.Println(err.Error())
		}
		sqlStmt.Exec(ml[i].User.ID, ml[i].User.Username, ml[i].User.Discriminator)
	}
}

func addNewUser(u *discordgo.User) {
	sqlStmt, err := DB.Prepare("INSERT INTO users (id, name, discrim, exp, vexp, wexp, wvexp, mexp, mvexp, aml, amlc) values (?, ?, ?, '0', '0', '0', '0', '0', '0', '0', '0');")
	if err != nil {
		fmt.Println(err.Error())
	}
	sqlStmt.Exec(u.ID, u.Username, u.Discriminator)
}

func addExp(m *discordgo.MessageCreate) {
	u := new(user)
	err := DB.QueryRow("SELECT * FROM users WHERE id = ?;", m.Author.ID).Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
	if err != nil {
		fmt.Println(err.Error())
	}
	sqlStmt, err := DB.Prepare("UPDATE users SET exp = ?, wexp = ?, mexp = ?, aml = ?, amlc = ? WHERE id = ?;")
	if err != nil {
		fmt.Println(err.Error())
	}
	u.exp = u.exp + 10
	u.wexp = u.wexp + 10
	u.mexp = u.mexp + 10
	if len(m.Content) > 0 {
		u.amlc = u.amlc + 1
		u.aml = u.aml + ((float32(len(m.Content)) - u.aml) / float32(u.amlc))
	}
	sqlStmt.Exec(u.exp, u.wexp, u.mexp, u.aml, u.amlc, m.Author.ID)
}

func printLeaderboard(s *discordgo.Session, m *discordgo.MessageCreate) {
	u := new(user)

	text := ""
	rows, _ := DB.Query("SELECT * FROM users ORDER BY exp DESC LIMIT 10;")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
		text = text + u.name + ": " + strconv.Itoa(u.exp) + "\n"
	}

	textVC := ""
	rowsVC, _ := DB.Query("SELECT * FROM users ORDER BY vexp DESC LIMIT 10;")
	defer rowsVC.Close()
	for rowsVC.Next() {
		rowsVC.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
		textVC = textVC + u.name + ": " + secsToHours(u.vexp) + "\n"
	}

	mE := new(discordgo.MessageEmbed)
	mE.Color = 9693630
	mE.Title = "Leaderboards"

	f1 := new(discordgo.MessageEmbedField)
	f1.Inline = true
	f1.Name = "Text Chat"
	f1.Value = text
	mE.Fields = append(mE.Fields, f1)

	f2 := new(discordgo.MessageEmbedField)
	f2.Inline = true
	f2.Name = "Voice Chat"
	f2.Value = textVC
	mE.Fields = append(mE.Fields, f2)

	_, err := s.ChannelMessageSendEmbed(BotCommandsChannel, mE)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func printWeeklyLeaderboard(s *discordgo.Session, m *discordgo.MessageCreate) {
	u := new(user)

	text := ""
	rows, _ := DB.Query("SELECT * FROM users ORDER BY wexp DESC LIMIT 10;")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
		text = text + u.name + ": " + strconv.Itoa(u.wexp) + "\n"
	}

	textVC := ""
	rowsVC, _ := DB.Query("SELECT * FROM users ORDER BY wvexp DESC LIMIT 10;")
	defer rowsVC.Close()
	for rowsVC.Next() {
		rowsVC.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
		textVC = textVC + u.name + ": " + secsToHours(u.wvexp) + "\n"
	}

	mE := new(discordgo.MessageEmbed)
	mE.Color = 9693630
	mE.Title = "Weekly Leaderboards"

	f1 := new(discordgo.MessageEmbedField)
	f1.Inline = true
	f1.Name = "Text Chat"
	f1.Value = text
	mE.Fields = append(mE.Fields, f1)

	f2 := new(discordgo.MessageEmbedField)
	f2.Inline = true
	f2.Name = "Voice Chat"
	f2.Value = textVC
	mE.Fields = append(mE.Fields, f2)

	_, err := s.ChannelMessageSendEmbed(BotCommandsChannel, mE)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func printMonthlyLeaderboard(s *discordgo.Session, m *discordgo.MessageCreate) {
	u := new(user)

	text := ""
	rows, _ := DB.Query("SELECT * FROM users ORDER BY mexp DESC LIMIT 10;")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
		text = text + u.name + ": " + strconv.Itoa(u.mexp) + "\n"
	}

	textVC := ""
	rowsVC, _ := DB.Query("SELECT * FROM users ORDER BY mvexp DESC LIMIT 10;")
	defer rowsVC.Close()
	for rowsVC.Next() {
		rowsVC.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
		textVC = textVC + u.name + ": " + secsToHours(u.mvexp) + "\n"
	}

	mE := new(discordgo.MessageEmbed)
	mE.Color = 9693630
	mE.Title = "Monthly Leaderboards"

	f1 := new(discordgo.MessageEmbedField)
	f1.Inline = true
	f1.Name = "Text Chat"
	f1.Value = text
	mE.Fields = append(mE.Fields, f1)

	f2 := new(discordgo.MessageEmbedField)
	f2.Inline = true
	f2.Name = "Voice Chat"
	f2.Value = textVC
	mE.Fields = append(mE.Fields, f2)

	_, err := s.ChannelMessageSendEmbed(BotCommandsChannel, mE)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func printTextLengthLeaderboard(s *discordgo.Session, m *discordgo.MessageCreate) {
	u := new(user)

	textTop := ""
	rows, _ := DB.Query("SELECT * FROM users ORDER BY aml DESC LIMIT 10;")
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
		textTop = textTop + u.name + ": " + fmt.Sprintf("%.2f", u.aml) + "\n"
	}

	textBot := ""
	rowsVC, _ := DB.Query("SELECT * FROM users WHERE aml > 0 ORDER BY aml ASC LIMIT 10;")
	defer rowsVC.Close()
	for rowsVC.Next() {
		rowsVC.Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
		textBot = textBot + u.name + ": " + fmt.Sprintf("%.2f", u.aml) + "\n"
	}

	mE := new(discordgo.MessageEmbed)
	mE.Color = 9693630
	mE.Title = "Monthly Leaderboards"

	f1 := new(discordgo.MessageEmbedField)
	f1.Inline = true
	f1.Name = "Ten Longest"
	f1.Value = textTop
	mE.Fields = append(mE.Fields, f1)

	f2 := new(discordgo.MessageEmbedField)
	f2.Inline = true
	f2.Name = "Ten Shortest"
	f2.Value = textBot
	mE.Fields = append(mE.Fields, f2)

	_, err := s.ChannelMessageSendEmbed(BotCommandsChannel, mE)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func findExp(m *discordgo.MessageCreate) (int, int, int, int, int, int) {
	u := new(user)
	err := DB.QueryRow("SELECT * FROM users WHERE id = ?;", m.Author.ID).Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
	if err != nil {
		fmt.Println(err.Error())
	}
	return u.exp, u.vexp, u.wexp, u.wvexp, u.mexp, u.mvexp
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
	err := DB.QueryRow("SELECT * FROM users WHERE id = ?;", m.UserID).Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
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

func clearWeeklyExp() {
	sqlStmt, err := DB.Prepare("UPDATE users SET wexp = 0, wvexp = 0;")
	if err != nil {
		fmt.Println(err.Error())
	}
	sqlStmt.Exec()
}

func clearMonthlyExp() {
	sqlStmt, err := DB.Prepare("UPDATE users SET mexp = 0, mvexp = 0;")
	if err != nil {
		fmt.Println(err.Error())
	}
	sqlStmt.Exec()
}

func getWeeklyExp(s *discordgo.Session) {
	u := new(user)
	message := "```\nMost active users last week:\n"
	err := DB.QueryRow("SELECT * FROM users ORDER BY wexp DESC LIMIT 1;").Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
	if err != nil {
		fmt.Println(err.Error())
	}
	message = message + "Text chat -> " + u.name + " with " + strconv.Itoa(u.wexp/10) + " messages sent!\n"
	err = DB.QueryRow("SELECT * FROM users ORDER BY wvexp DESC LIMIT 1;").Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
	if err != nil {
		fmt.Println(err.Error())
	}
	message = message + "Voice chat -> " + u.name + " with " + secsToHours(u.wvexp) + " spent in chat!\n```"
	_, _ = s.ChannelMessageSend(HallOfFameChannel, message)
}

func getMonthlyExp(s *discordgo.Session) {
	u := new(user)
	message := "**```\nMost active users last month:\n"
	err := DB.QueryRow("SELECT * FROM users ORDER BY mexp DESC LIMIT 1;").Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
	if err != nil {
		fmt.Println(err.Error())
	}
	message = message + "Text chat -> " + u.name + " with " + strconv.Itoa(u.mexp/10) + " messages sent!\n"
	err = DB.QueryRow("SELECT * FROM users ORDER BY mvexp DESC LIMIT 1;").Scan(&u.id, &u.name, &u.discrim, &u.exp, &u.vexp, &u.wexp, &u.wvexp, &u.mexp, &u.mvexp, &u.aml, &u.amlc)
	if err != nil {
		fmt.Println(err.Error())
	}
	message = message + "Voice chat -> " + u.name + " with " + secsToHours(u.mvexp) + " spent in chat!\n```**"
	_, _ = s.ChannelMessageSend(HallOfFameChannel, message)
}

func findWeeklyPos(m *discordgo.MessageCreate, wexp int) int {
	var ranking int
	err := DB.QueryRow("SELECT COUNT (*) FROM users WHERE wexp >= ?;", wexp).Scan(&ranking)
	if err != nil {
		fmt.Println(err.Error())
	}
	return ranking
}

func findWeeklyVCPos(m *discordgo.MessageCreate, wvexp int) int {
	var ranking int
	err := DB.QueryRow("SELECT COUNT (*) FROM users WHERE wvexp >= ?;", wvexp).Scan(&ranking)
	if err != nil {
		fmt.Println(err.Error())
	}
	return ranking
}
func findMonthlyPos(m *discordgo.MessageCreate, mexp int) int {
	var ranking int
	err := DB.QueryRow("SELECT COUNT (*) FROM users WHERE mexp >= ?;", mexp).Scan(&ranking)
	if err != nil {
		fmt.Println(err.Error())
	}
	return ranking
}

func findMonthlyVCPos(m *discordgo.MessageCreate, mvexp int) int {
	var ranking int
	err := DB.QueryRow("SELECT COUNT (*) FROM users WHERE mvexp >= ?;", mvexp).Scan(&ranking)
	if err != nil {
		fmt.Println(err.Error())
	}
	return ranking
}
