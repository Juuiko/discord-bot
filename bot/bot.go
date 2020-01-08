package bot

import (
   "../config"
   "fmt"
   "github.com/bwmarrin/discordgo"
   "strings"
   "math/rand"
   "time"
   "io/ioutil"
   "net/http"
   "strconv"
)

var BotID string
var goBot *discordgo.Session
var vcCon *discordgo.VoiceConnection

func userGoodbye(s *discordgo.Session, u *discordgo.GuildMemberRemove){
   _, _ = s.ChannelMessageSend(WelcomeChannel,  fmt.Sprintf("%s was banned by the tyranical Crassus, Fs in chat pls <:OBKick:643516408994594817> :cry:", u.User.Username))
   return
}

func userWelcome(s *discordgo.Session, u *discordgo.GuildMemberAdd){
   _, _ = s.ChannelMessageSend(WelcomeChannel, fmt.Sprintf("Hey %s, welcome to **Quantex Esports Network** :tada: :hugging: <:OBKiss:643520085197062164> !", u.User.Mention()))
   return
}

func logHandler(s *discordgo.Session, m *discordgo.MessageCreate){
   if m.Author.ID == BotID {
      return
   }
   if strings.Contains(m.Content, "@") {m.Content = strings.Replace(m.Content, "@", "[at]", 10)}
   _, _ = s.ChannelMessageSend(LogsChannel, fmt.Sprintf("\"%s\" - %s in %s", m.Content, m.Author.String(), ChannelNameByID[m.ChannelID]))
}

func profileEmbed(s *discordgo.Session, m *discordgo.MessageCreate){
   mE := new(discordgo.MessageEmbed)
   mE.Title = fmt.Sprintf("%s's profile",m.Author.Username)
   pic := new(discordgo.MessageEmbedImage)
   pic.URL = m.Author.AvatarURL("128")
   mE.Image = pic
   mE.Color = 9693630
   exp := findExp(m)
   pos := findPos(m,exp)
   member, err := s.GuildMember(QuantexID,m.Author.ID)
   if err != nil {
    fmt.Println(err.Error())
   }
   time, err := member.JoinedAt.Parse()
   if err != nil {
    fmt.Println(err.Error())
   }
   mE.Description = "Server exp = " + strconv.Itoa(exp) + "\n Sever rank = " + strconv.Itoa(pos) + "\nJoin date = " + time.Format("02/01/2006 15:04")
   _, err = s.ChannelMessageSendEmbed(BotCommandsChannel,mE)
   if err != nil {
    fmt.Println(err.Error())
   }
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate){
   if m.Author.ID == BotID {
      return
   }
   addExp(m)
   if m.ChannelID == BotTestChannel {
      if strings.HasPrefix(m.Content, config.BotPrefix) {
         switch m.Content {
         case "!listUsers":
            memberListArray, _ := s.GuildMembers(QuantexID,"90530967382417408", 1000)
            userList := "Total user list:\n"
            for i := 0; i<len(memberListArray); i++ {
               userList = userList + memberListArray[i].User.Username + "\n"
            }
            _, _ = s.ChannelMessageSend(BotTestChannel, userList)
         case "!countUsers":
            memberListArray, _ := s.GuildMembers(QuantexID,"0", 1000)
            _, _ = s.ChannelMessageSend(BotTestChannel, fmt.Sprintf("Number of users: %d", strconv.Itoa(len(memberListArray)) ))
         case "!updateList":
            memberListArray, _ := s.GuildMembers(QuantexID,"0",1000)
            fillDB(memberListArray)
         }
      }
   } else if m.ChannelID == BotCommandsChannel || m.ChannelID == BotTestChannel {
      if strings.HasPrefix(m.Content, config.BotPrefix) {
         switch m.Content {
         case "!help":
            _, _ = s.ChannelMessageSend(BotCommandsChannel, "Command list: cointoss, ping, inspire, join, exit, top, me")
         case "!cointoss":
            command_cointoss(s,m)
         case "!top":
            printLeaderboard(s,m)
        case "!me":
            profileEmbed(s,m)
         case "!ping":
            _, _ = s.ChannelMessageSend(BotCommandsChannel, "pong")
         case "!emoteT":
            _, _ = s.ChannelMessageSend(BotCommandsChannel, "<:OBKiss:643520085197062164>")
         case "!inspire":
            command_inspire(s,m)
         case "!join":
            vcCon, _ = s.ChannelVoiceJoin(QuantexID,"591762237857726484",false, false)
         case "!exit":
            vcCon.Disconnect()
         }
      }
   }
}

   func command_inspire(s *discordgo.Session, m *discordgo.MessageCreate){
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

   func command_cointoss(s *discordgo.Session, m *discordgo.MessageCreate){
      coin := []string{
              "heads",
              "tails",
      }
      rand.Seed(time.Now().UnixNano())
      side := coin[rand.Intn(len(coin))]
      _, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("The coin landed on %s!", side))
   }

func Start() {
   goBot , err := discordgo.New("Bot " + config.Token)
   if err!= nil {
      fmt.Println(err.Error())
      return
   }

   u, err := goBot.User("@me")
   if err!= nil {
      fmt.Println(err.Error())
   }

   BotID = u.ID
   goBot.AddHandler(userGoodbye)
   goBot.AddHandler(userWelcome)
   goBot.AddHandler(logHandler)
   goBot.AddHandler(messageHandler)

   err = goBot.Open()
   if err!= nil {
      fmt.Println(err.Error())
      return
   }

   fmt.Println("Bot is running!")

}
