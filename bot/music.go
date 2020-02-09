package bot

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
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
	duration time.Duration
}

var q queue

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
		_, _ = s.ChannelMessageSend(BotCommandsChannel, "You have to be in a voice channel to use this command")
	} else {
		vcCon, _ = s.ChannelVoiceJoin(QuantexID, channel, false, false)
	}
}

func musicCommandLeave() {
	vcCon.Disconnect()
	q.running = false
	//q.list = q.list[:0]
}

func musicCommandQueue(s *discordgo.Session) {
	queueString := ""
	if len(q.list) < 1 {
		queueString = "The queue is empty!"
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

func musicCommandPlay(s *discordgo.Session) {
	if q.running {
		_, _ = s.ChannelMessageSend(BotCommandsChannel, "Already running!")
	} else {
		q.running = true
		for len(q.list) > 0 {
			_, _ = s.ChannelMessageSend(BotCommandsChannel, fmt.Sprintf("```Now playing: %s```", q.list[0].title))
			play(vcCon, "./music/"+q.list[0].title, make(chan bool))
			var err = os.Remove("./music/" + q.list[0].title)
			if err != nil {
				fmt.Println(err)
			}
			q.list = q.list[1:]
		}
		q.running = false
	}
}

func addMusic(s *discordgo.Session, m *discordgo.MessageCreate) {
	songURL := strings.Trim(m.Content, "!addSong ")
	vid, err := ytdl.GetVideoInfo(songURL)
	if err != nil {
		fmt.Println("Failed to get video info ->", err)
	}
	newSong := song{
		title:    vid.Title,
		url:      songURL,
		duration: vid.Duration,
	}
	q.list = append(q.list, newSong)
	fmt.Println(q)
	fmt.Println(vid.Title)
	file, _ := os.Create("./music/" + vid.Title)
	defer file.Close()
	vid.Download(vid.Formats[0], file)
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
