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
   _, _ = s.ChannelMessageSend(WelcomeChannel,  u.User.Username + " was banned by the tyranical Crassus, Fs in chat pls <:OBKick:643516408994594817> :cry:")
   return
}

func userWelcome(s *discordgo.Session, u *discordgo.GuildMemberAdd){
   _, _ = s.ChannelMessageSend(WelcomeChannel, "Hey " + u.User.Mention() + ", welcome to **Quantex Esports Network** :tada: :hugging: <:OBKiss:643520085197062164> !")
   return
}

func logHandler(s *discordgo.Session, m *discordgo.MessageCreate){
   if m.Author.ID == BotID {
      return
   }
   _, _ = s.ChannelMessageSend(LogsChannel, "\"" + m.Content + "\" - " + m.Author.String() + " in " + ChannelNameByID[m.ChannelID])
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate){
   if m.Author.ID == BotID {
      return
   }
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
            _, _ = s.ChannelMessageSend(BotTestChannel, "Number of users: " + strconv.Itoa(len(memberListArray)))
         }
      }
   } else if m.ChannelID == BotCommandsChannel || m.ChannelID == BotTestChannel {
      if strings.HasPrefix(m.Content, config.BotPrefix) {
         switch m.Content {
         case "!help":
            _, _ = s.ChannelMessageSend(BotCommandsChannel, "Command list: cointoss, ping, inspire, join, exit")
         case "!cointoss":
            command_cointoss(s,m)
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
      _, _ = s.ChannelMessageSend(m.ChannelID, "The coin landed on " + side +"!")
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
