package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"../config"
	"github.com/bwmarrin/discordgo"
	"github.com/jasonlvhit/gocron"
)

// BotID is id for bot
var BotID string
var goBot *discordgo.Session
var vcCon *discordgo.VoiceConnection
var leagueAPIBusy bool

// ConnectionMap is all users currently in a voice chat
var ConnectionMap map[string]int64

func userGoodbye(s *discordgo.Session, u *discordgo.GuildMemberRemove) {
	_, _ = s.ChannelMessageSend(WelcomeChannel, fmt.Sprintf("%s was banned by the tyranical Crassus, Fs in chat pls <:OBKick:643516408994594817> :cry:", u.User.Username))
	return
}

func userWelcome(s *discordgo.Session, u *discordgo.GuildMemberAdd) {
	_, _ = s.ChannelMessageSend(WelcomeChannel, fmt.Sprintf("Hey %s, welcome to **Quantex Esports Network** :tada: :hugging: <:OBKiss:643520085197062164> !", u.User.Mention()))
	addNewUser(u.User)
	m, _ := s.ChannelMessageSend(RoleSelectChannel, u.User.Mention())
	_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
	return
}

func logHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotID {
		return
	}
	if strings.Contains(m.Content, "@") {
		m.Content = strings.Replace(m.Content, "@", "[at]", 10)
	}
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = s.ChannelMessageSend(LogsChannel, fmt.Sprintf("\"%s\" - %s in %s", m.Content, m.Author.String(), channel.Name))
	if err != nil {
		fmt.Println(err.Error())
	}
}

func voiceHandler(s *discordgo.Session, u *discordgo.VoiceStateUpdate) {
	_, inMap := ConnectionMap[u.UserID]
	if u.UserID == BotID {
		return
	} else if u.ChannelID == "" || u.Suppress || u.SelfMute || u.SelfDeaf || u.Mute || u.Deaf {
		if !inMap {
			return
		}
		addTimeToDB(time.Now().Unix()-ConnectionMap[u.UserID], u)
		delete(ConnectionMap, u.UserID)
	} else {
		if !inMap {
			ConnectionMap[u.UserID] = time.Now().Unix()
		}
	}
}

func secsToHours(secs int) string {
	var str1, str2 string
	mins := secs / 60
	hours := mins / 60
	minsLeftover := mins % 60
	if hours == 1 {
		str1 = fmt.Sprintf("%dhr ", hours)
	} else {
		str1 = fmt.Sprintf("%dhrs ", hours)
	}
	if minsLeftover == 1 {
		str2 = fmt.Sprintf("%dmin", minsLeftover)
	} else {
		str2 = fmt.Sprintf("%dmins", minsLeftover)
	}
	str := str1 + str2
	return str
}

func profileEmbed(s *discordgo.Session, m *discordgo.MessageCreate) {
	exp, vexp, wexp, wvexp, mexp, mvexp, aml := findExp(m)
	pos := findPos(m, exp)
	vcPos := findVCPos(m, vexp)
	wpos := findWeeklyPos(m, wexp)
	wVCPos := findWeeklyVCPos(m, wvexp)
	mpos := findMonthlyPos(m, mexp)
	mVCPos := findMonthlyVCPos(m, mvexp)
	amlPos := findAMLPos(m, aml)

	mE := new(discordgo.MessageEmbed)
	mE.Color = 9693630
	mE.Description = fmt.Sprintf("Chat exp = %v\nChat rank = %v\nVC time = %v\nVoice rank = %v\nAvg Msg Length = %v\nAML rank = %v", exp, pos, secsToHours(vexp), vcPos, aml, amlPos)

	author := new(discordgo.MessageEmbedAuthor)
	author.Name = fmt.Sprintf("%s's profile", m.Author.Username)
	author.IconURL = m.Author.AvatarURL("128")
	mE.Author = author

	footer := new(discordgo.MessageEmbedFooter)
	member, _ := s.GuildMember(QuantexID, m.Author.ID)
	time, _ := member.JoinedAt.Parse()
	footer.Text = fmt.Sprintf("Joined on %v", time.Format("02/01/2006 15:04"))
	mE.Footer = footer

	f1 := new(discordgo.MessageEmbedField)
	f1.Inline = true
	f1.Name = "Weekly"
	f1.Value = fmt.Sprintf("Chat exp = %v\nChat rank = %v\nVC time = %v\nVoice rank = %v", wexp, wpos, secsToHours(wvexp), wVCPos)
	mE.Fields = append(mE.Fields, f1)

	f2 := new(discordgo.MessageEmbedField)
	f2.Inline = true
	f2.Name = "Monthly"
	f2.Value = fmt.Sprintf("Chat exp = %v\nChat rank = %v\nVC time = %v\nVoice rank = %v", mexp, mpos, secsToHours(mvexp), mVCPos)
	mE.Fields = append(mE.Fields, f2)

	_, err3 := s.ChannelMessageSendEmbed(BotCommandsChannel, mE)
	if err3 != nil {
		fmt.Println(err3.Error())
	}
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotID {
		return
	}
	addExp(m)
	if strings.ToLower(m.Content) == "who asked?" || strings.ToLower(m.Content) == "who asked" {
		rand.Seed(time.Now().UnixNano())
		if rand.Intn(100) == 0 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "__***I asked.***__")
		}
	}
	if !strings.HasPrefix(m.Content, "!addSong") {
		m.Content = strings.ToLower(m.Content)
	}
	if strings.HasPrefix(m.Content, "!insecure") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=4PG_elEG7rA")
	} else if strings.HasPrefix(m.Content, "!gif") {
		commandGiphy(s, m)
	} else if strings.HasPrefix(m.Content, "!lolstats") {
		m.Content = m.Content[10:]
		commandLeagueStats(s, m)
	} else if strings.HasPrefix(m.Content, "!bettertop") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=C2iK35Mtgbk")
	} else if strings.HasPrefix(m.Content, "!betterjungle") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=D8IjiKj-U5c")
	} else if strings.HasPrefix(m.Content, "!bettermid") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=3aUa_xVjf-w")
	} else if strings.HasPrefix(m.Content, "!betterbot") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=coJJoFdIitM")
	} else if strings.HasPrefix(m.Content, "!bettersupport") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=ivWbsc4pGUc")
	} else if m.ChannelID == BotCommandsChannel {
		if strings.HasPrefix(m.Content, config.BotPrefix) {
			if strings.HasPrefix(m.Content, "!addSong") {
				addMusic(s, m)
			} else {
				switch m.Content {
				case "!help":
					_, _ = s.ChannelMessageSend(BotCommandsChannel, "```Command list: cointoss, inspire, top, topWeek, topMonth, avgText, topEgirls, addSong, play, skip, clearQueue, better*Role*, insecure, me```")
				case "!cointoss":
					commandCointoss(s, m)
				case "!top":
					printLeaderboard(s, m)
				case "!topweek":
					printWeeklyLeaderboard(s, m)
				case "!topmonth":
					printMonthlyLeaderboard(s, m)
				case "!avgtext":
					printTextLengthLeaderboard(s, m)
				case "!topegirls":
					_, _ = s.ChannelMessageSend(BotCommandsChannel, "```Top Quantex Egirls:\n1. Neasa\n2. bgscurtis\n3. Raj\n4. Adam\n5. Lizzy```")
				case "!me":
					profileEmbed(s, m)
				case "!inspire":
					commandInspire(s, m)
				case "!skip":
					musicCommandSkip(s)
				case "!clearqueue":
					musicCommandClearQueue(s)
				case "!play":
					musicCommandPlay(s, m)
				case "!queue":
					musicCommandQueue(s)
				default:
					_, _ = s.ChannelMessageSend(BotCommandsChannel, "```Invalid command, !help for the command list```")
				}
			}
		}
	}
}

func commandGiphy(s *discordgo.Session, m *discordgo.MessageCreate) {
	input := strings.Trim(m.Content, "!gif ")
	input = strings.Replace(input, " ", "%", 100)
	url := "https://api.giphy.com/v1/gifs/search?api_key=" + config.GiphyKey + "&q=" + input + "&limit=1"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	var data map[string]interface{}
	err = json.Unmarshal([]byte(html), &data)
	if err != nil {
		fmt.Println(err.Error())
	}
	if data["meta"].(map[string]interface{})["status"].(float64) == 200 {
		if len(data["data"].([]interface{})) == 0 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "```No results found :(```")
		} else {
			gif := data["data"].([]interface{})[0].(map[string]interface{})["embed_url"]
			_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s", gif))
		}
	} else {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("API error -> %s", data["meta"].(map[string]interface{})["msg"]))
	}
}

func commandInspire(s *discordgo.Session, m *discordgo.MessageCreate) {
	url := "https://inspirobot.me/api?generate=true"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, string(html))
}

func commandLeagueStats(s *discordgo.Session, m *discordgo.MessageCreate) {
	if leagueAPIBusy == false {
		leagueAPIBusy = true
		mE := new(discordgo.MessageEmbed)
		mE.Color = 9693630
		mE.Title = "League API is Skooking Away..."
		mE.Image = &discordgo.MessageEmbedImage{
			URL: "https://i.gifer.com/Sge7.gif",
		}
		loading, _ := s.ChannelMessageSendEmbed(m.ChannelID, mE)
		name := m.Content
		err := makeGraph(name)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "```Summoner name could not be found on EUW!```")
			fmt.Println(err)
		} else {
			filename := "./barchart.png"
			f, err := os.Open(filename)
			if err != nil {
				fmt.Println(err.Error())
			}
			defer f.Close()
			message := &discordgo.MessageSend{
				Content: fmt.Sprintf("**%v's League Stats!**", name),
				Files: []*discordgo.File{
					&discordgo.File{
						Name:   filename,
						Reader: f,
					},
				},
			}
			s.ChannelMessageSendComplex(m.ChannelID, message)
		}
		mE.Title = "League API Cooling Down..."
		loading, _ = s.ChannelMessageEditEmbed(loading.ChannelID, loading.ID, mE)
		time.Sleep(2 * time.Minute)
		_ = s.ChannelMessageDelete(loading.ChannelID, loading.ID)
		leagueAPIBusy = false
	} else {
		_, _ = s.ChannelMessageSend(m.ChannelID, "```The League API is busy atm, try again later!```")
	}
}

func commandCointoss(s *discordgo.Session, m *discordgo.MessageCreate) {
	coin := []string{
		"heads",
		"tails",
	}
	rand.Seed(time.Now().UnixNano())
	side := coin[rand.Intn(len(coin))]
	_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```The coin landed on %s!```", side))
}

func messageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.MessageReaction.MessageID == "674756079002976268" {
		if r.MessageReaction.Emoji.ID == "674659046468354057" { //EU4
			s.GuildMemberRoleAdd(QuantexID, r.UserID, "669343691655348236")
		}
		if r.MessageReaction.Emoji.ID == "674659075560308766" { //LoL
			s.GuildMemberRoleAdd(QuantexID, r.UserID, "669344064780632175")
		}
		if r.MessageReaction.Emoji.ID == "674659105578811402" { //Minecraft
			s.GuildMemberRoleAdd(QuantexID, r.UserID, "669344100100603916")
		}
		if r.MessageReaction.Emoji.ID == "701157578046046218" { //Civ
			s.GuildMemberRoleAdd(QuantexID, r.UserID, "701158068326498325")
		}
		if r.MessageReaction.Emoji.Name == "🇨" { //Customs
			s.GuildMemberRoleAdd(QuantexID, r.UserID, "680186040052613176")
		}
		if r.MessageReaction.Emoji.ID == "704239699086016562" { //Valorant
			s.GuildMemberRoleAdd(QuantexID, r.UserID, "704238882710880336")
		}
		if r.MessageReaction.Emoji.ID == "747604229014814760" { //Among Us
			s.GuildMemberRoleAdd(QuantexID, r.UserID, "761327960853577738")
		}
	}
}

func messageReactionDel(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	if r.MessageReaction.MessageID == "674756079002976268" {
		if r.MessageReaction.Emoji.ID == "674659046468354057" {
			s.GuildMemberRoleRemove(QuantexID, r.UserID, "669343691655348236")
		}
		if r.MessageReaction.Emoji.ID == "674659075560308766" {
			s.GuildMemberRoleRemove(QuantexID, r.UserID, "669344064780632175")
		}
		if r.MessageReaction.Emoji.ID == "674659105578811402" {
			s.GuildMemberRoleRemove(QuantexID, r.UserID, "669344100100603916")
		}
		if r.MessageReaction.Emoji.ID == "701157578046046218" {
			s.GuildMemberRoleRemove(QuantexID, r.UserID, "701158068326498325")
		}
		if r.MessageReaction.Emoji.Name == "🇨" {
			s.GuildMemberRoleRemove(QuantexID, r.UserID, "680186040052613176")
		}
		if r.MessageReaction.Emoji.ID == "704239699086016562" {
			s.GuildMemberRoleRemove(QuantexID, r.UserID, "704238882710880336")
		}
		if r.MessageReaction.Emoji.ID == "747604229014814760" {
			s.GuildMemberRoleRemove(QuantexID, r.UserID, "761327960853577738")
		}
	}
}

func task(s *discordgo.Session) {
	t := time.Now()
	if t.Weekday() == 1 {
		getWeeklyExp(s)
		clearWeeklyExp()
		updateUserNames(s)
	}
	if t.Day() == 1 {
		getMonthlyExp(s)
		clearMonthlyExp()
	}
}

// Start is bot keep awake function
func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	u, err := goBot.User("@me")
	if err != nil {
		fmt.Println(err.Error())
	}

	BotID = u.ID
	goBot.AddHandler(userGoodbye)
	goBot.AddHandler(userWelcome)
	goBot.AddHandler(logHandler)
	goBot.AddHandler(messageHandler)
	goBot.AddHandler(voiceHandler)
	goBot.AddHandler(messageReactionAdd)
	goBot.AddHandler(messageReactionDel)

	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ConnectionMap = make(map[string]int64)
	fmt.Println("Bot is running!")
	gocron.Every(1).Day().At("00:00").Do(task, goBot)
	<-gocron.Start()
}
