package bot

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"net/http"
	"os/exec"
	"strconv"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/google/uuid"
	"github.com/rylio/ytdl"
	"layeh.com/gopus"
)

type queue struct {
	list    []song
	running bool
}

type song struct {
	title    string
	url      string
	filename string
}

var q queue
var stopPlayback chan bool

func getVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate) string {
	guildObj, _ := s.Guild(m.GuildID)
	for i := 0; i < len(guildObj.VoiceStates); i++ {
		if guildObj.VoiceStates[i].UserID == m.Author.ID {
			return guildObj.VoiceStates[i].ChannelID
		}
	}
	return ""
}

func musicCommandJoin(s *discordgo.Session, m *discordgo.MessageCreate) {
	channel := getVoiceChannel(s, m)
	if channel == "" {
		_, _ = s.ChannelMessageSend(BotCommandsChannel, "```You have to be in a voice channel to use this command```")
	} else {
		vcCon, _ = s.ChannelVoiceJoin(QuantexID, channel, false, false)
	}
}

func musicCommandClearQueue(s *discordgo.Session) {
	if len(q.list) > 1 {
		for i := len(q.list) - 1; i > 0; i-- {
			var err = os.Remove("./music/" + q.list[i].filename)
			if err != nil {
				fmt.Println(err)
			}
		}
		q.list = append(q.list[:1])
		_, _ = s.ChannelMessageSend(BotCommandsChannel, "```Queue cleared!```")
	} else {
		_, _ = s.ChannelMessageSend(BotCommandsChannel, "```Queue needs to be greater than 1 for this command```")
	}
}

func musicCommandLeave() {
	vcCon.Disconnect()
	if q.running {
		var err = os.Remove("./music/" + q.list[0].filename)
		if err != nil {
			fmt.Println(err)
		}
		q.list = q.list[1:]
		q.running = false
	}
}

func musicCommandSkip(s *discordgo.Session) {
	if !q.running {
		_, _ = s.ChannelMessageSend(BotCommandsChannel, "```Skipped "+q.list[0].title+"!```")
	} else {
		_, _ = s.ChannelMessageSend(BotCommandsChannel, "```Skipped "+q.list[0].title+"!```")
		stopPlayback <- true
	}
}

func musicCommandQueue(s *discordgo.Session) {
	queueString := ""
	if len(q.list) < 1 {
		queueString = "```The queue is empty```"
		_, _ = s.ChannelMessageSend(BotCommandsChannel, queueString)
	} else {
		queueString = "```Music Queue:\n"
		for i := 0; i < len(q.list); i++ {
			queueString = queueString + strconv.Itoa(i+1) + ". " + q.list[i].title + "\n"
		}
		queueString += "```"
		_, _ = s.ChannelMessageSend(BotCommandsChannel, queueString)
	}
}

func musicCommandPlay(s *discordgo.Session, m *discordgo.MessageCreate) {
	if q.running {
		_, _ = s.ChannelMessageSend(BotCommandsChannel, "```Already running```")
	} else if len(q.list) <= 0 {
		_, _ = s.ChannelMessageSend(BotCommandsChannel, "```Empty queue, try adding some songs first```")
	} else {
		musicCommandJoin(s, m)
		q.running = true
		stopPlayback = make(chan bool)
		for len(q.list) > 0 {
			_, _ = s.ChannelMessageSend(BotCommandsChannel, fmt.Sprintf("```Now playing: %s```", q.list[0].title))
			play(vcCon, "./music/"+q.list[0].filename, stopPlayback)
			var err = os.Remove("./music/" + q.list[0].filename)
			if err != nil {
				fmt.Println(err)
			}
			q.list = q.list[1:]
		}
		q.running = false
		musicCommandLeave()
		_, _ = s.ChannelMessageSend(BotCommandsChannel, "```Queue reached end, goodbye!```")
	}
}

func addMusic(s *discordgo.Session, m *discordgo.MessageCreate) {
	songURL := m.Content[len(m.Content)-11:]
	c := ytdl.Client{
		HTTPClient: http.DefaultClient,
		Logger:     log.Logger,
	}
	vid, err := c.GetVideoInfoFromID(songURL)
	if err != nil {
		_, _ = s.ChannelMessageSend(BotCommandsChannel, fmt.Sprintf("```Failed to get video info -> %s```", err))
		return
	}
	newSong := song{
		title:    vid.Title,
		url:      songURL,
		filename: uuid.New().String(),
	}
	q.list = append(q.list, newSong)
	_, _ = s.ChannelMessageSend(BotCommandsChannel, fmt.Sprintf("```Added %s to queue```", newSong.title))
	file, _ := os.Create("./music/" + newSong.filename)
	defer file.Close()
	c.Download(vid, vid.Formats[0], file)
}

//*****Thanks to bwmarrin for the file to audio code!! <3********//

const (
	channels  int = 2                   // 1 for mono, 2 for stereo
	frameRate int = 48000               // audio sampling rate
	frameSize int = 960                 // uint16 size of each audio frame
	maxBytes  int = (frameSize * 2) * 2 // max size of opus data
)

var (
	speakers    map[uint32]*gopus.Decoder
	opusEncoder *gopus.Encoder
	mu          sync.Mutex
)

// SendPCM will receive on the provied channel encode
// received PCM data into Opus then send that to Discordgo
func SendPCM(v *discordgo.VoiceConnection, pcm <-chan []int16) {
	if pcm == nil {
		return
	}

	var err error

	opusEncoder, err = gopus.NewEncoder(frameRate, channels, gopus.Audio)

	if err != nil {
		fmt.Println(err)
		return
	}

	for {

		// read pcm from chan, exit if channel is closed.
		recv, ok := <-pcm
		if !ok {
			fmt.Println(err)
			return
		}

		// try encoding pcm frame with Opus
		opus, err := opusEncoder.Encode(recv, frameSize, maxBytes)
		if err != nil {
			fmt.Println(err)
			return
		}

		if v.Ready == false || v.OpusSend == nil {
			// OnError(fmt.Sprintf("Discordgo not ready for opus packets. %+v : %+v", v.Ready, v.OpusSend), nil)
			// Sending errors here might not be suited
			return
		}
		// send encoded opus data to the sendOpus channel
		v.OpusSend <- opus
	}
}

func play(v *discordgo.VoiceConnection, filename string, stop <-chan bool) {

	// Create a shell command "object" to run.
	run := exec.Command("ffmpeg", "-i", filename, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	ffmpegout, err := run.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	ffmpegbuf := bufio.NewReaderSize(ffmpegout, 16384)

	// Starts the ffmpeg command
	err = run.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		<-stop
		err = run.Process.Kill()
	}()

	// Send "speaking" packet over the voice websocket
	err = v.Speaking(true)
	if err != nil {
		fmt.Println(err)
	}

	// Send not "speaking" packet over the websocket when we finish
	defer func() {
		err := v.Speaking(false)
		if err != nil {
			fmt.Println(err)
		}
	}()

	send := make(chan []int16, 2)
	defer close(send)

	close := make(chan bool)
	go func() {
		SendPCM(v, send)
		close <- true
	}()

	for {
		// read data from ffmpeg stdout
		audiobuf := make([]int16, frameSize*channels)
		err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		// Send received PCM to the sendPCM channel
		select {
		case send <- audiobuf:
		case <-close:
			return
		}
	}
}
