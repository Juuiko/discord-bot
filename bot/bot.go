package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"../config"
	"github.com/bwmarrin/discordgo"
)

var BotID string
var goBot *discordgo.Session
var vcCon *discordgo.VoiceConnection

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
	channel, _ := s.Channel(m.ChannelID)
	_, _ = s.ChannelMessageSend(LogsChannel, fmt.Sprintf("\"%s\" - %s in %s", m.Content, m.Author.String(), channel.Name))
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
	mE := new(discordgo.MessageEmbed)
	mE.Title = fmt.Sprintf("%s's profile", m.Author.Username)

	pic := new(discordgo.MessageEmbedImage)
	pic.URL = m.Author.AvatarURL("128")
	mE.Image = pic
	mE.Color = 9693630
	exp, vexp := findExp(m)
	pos := findPos(m, exp)
	vcPos := findVCPos(m, vexp)
	member, err := s.GuildMember(QuantexID, m.Author.ID)
	if err != nil {
		fmt.Println(err.Error())
	}
	time, err := member.JoinedAt.Parse()
	if err != nil {
		fmt.Println(err.Error())
	}
	mE.Description = "Server exp = " + strconv.Itoa(exp) + "\nChat rank = " + strconv.Itoa(pos) + "\nVC time = " + secsToHours(vexp) + "\nVoice rank = " + strconv.Itoa(vcPos) + "\nJoin date = " + time.Format("02/01/2006 15:04")
	_, err = s.ChannelMessageSendEmbed(BotCommandsChannel, mE)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotID {
		return
	}
	addExp(m)
	if strings.HasPrefix(m.Content, "!insecure") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=4PG_elEG7rA")
	} else if strings.HasPrefix(m.Content, "!gif") {
		commandGiphy(s, m)
	} else if strings.HasPrefix(m.Content, "!betterTop") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=C2iK35Mtgbk")
	} else if strings.HasPrefix(m.Content, "!betterJungle") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=D8IjiKj-U5c")
	} else if strings.HasPrefix(m.Content, "!betterMid") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=3aUa_xVjf-w")
	} else if strings.HasPrefix(m.Content, "!betterBot") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=coJJoFdIitM")
	} else if strings.HasPrefix(m.Content, "!betterSupport") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "https://www.youtube.com/watch?v=ivWbsc4pGUc")
	} else if m.ChannelID == BotTestChannel {
		if strings.HasPrefix(m.Content, config.BotPrefix) {
			switch m.Content {
			case "!listUsers":
				memberListArray, _ := s.GuildMembers(QuantexID, "90530967382417408", 1000)
				userList := "Total user list:\n"
				for i := 0; i < len(memberListArray); i++ {
					userList = userList + memberListArray[i].User.Username + "\n"
				}
				_, _ = s.ChannelMessageSend(BotTestChannel, userList)
			case "!countUsers":
				memberListArray, _ := s.GuildMembers(QuantexID, "0", 1000)
				_, _ = s.ChannelMessageSend(BotTestChannel, fmt.Sprintf("Number of users: %s", strconv.Itoa(len(memberListArray))))
			case "!updateList":
				memberListArray, _ := s.GuildMembers(QuantexID, "0", 1000)
				fillDB(memberListArray)
			}
		}
	} else if m.ChannelID == BotCommandsChannel || m.ChannelID == BotTestChannel {
		if strings.HasPrefix(m.Content, config.BotPrefix) {
			if strings.HasPrefix(m.Content, "!calc") {
				result := strings.Split(m.Content, " ")
				output, err := calculate(result)
				if err != nil {
					_, _ = s.ChannelMessageSend(BotCommandsChannel, fmt.Sprintf(err.Error()))
				} else {
					_, _ = s.ChannelMessageSend(BotCommandsChannel, fmt.Sprintf("Answer -> %d", output))
				}
			} else if strings.HasPrefix(m.Content, "!addSong") {
				addMusic(s, m)
			} else {
				switch m.Content {
				case "!help":
					_, _ = s.ChannelMessageSend(BotCommandsChannel, "```Command list: cointoss, ping, inspire, top, topVC, topEgirls, calc, addSong, play, skip, clearQueue, insecure, me```")
				case "!cointoss":
					commandCointoss(s, m)
				case "!top":
					printLeaderboard(s, m)
				case "!topVC":
					printVCLeaderboard(s, m)
				case "!topEgirls":
					_, _ = s.ChannelMessageSend(BotCommandsChannel, "```Top Quantex Egirls:\n1. Neasa\n2. bgscurtis\n3. Raj\n4. Adam\n5. Lizzy```")
				case "!me":
					profileEmbed(s, m)
				case "!ping":
					_, _ = s.ChannelMessageSend(BotCommandsChannel, "```pong```")
				case "!emoteT":
					_, _ = s.ChannelMessageSend(BotCommandsChannel, "<:OBKiss:643520085197062164>")
				case "!inspire":
					commandInspire(s, m)
				case "!skip":
					musicCommandSkip(s)
				case "!clearQueue":
					musicCommandClearQueue(s)
				case "!play":
					musicCommandPlay(s, m)
				case "!queue":
					musicCommandQueue(s)
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
	}
}

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
}
