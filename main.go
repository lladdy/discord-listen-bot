package main

import (
	"flag"
	"fmt"
	"github.com/hajimehoshi/oto"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	//"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

// global audio context
var c *oto.Context
var player *oto.Player

func main() {
	flag.Parse()

	// Connect to Discord
	discord, err := discordgo.New("Bot " + "NjQ0NjkzMjczODg1MDE2MTA0.Xc3vcQ.4ujavEnx1H8L1S5xLpBxKvxg_l8")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Open Websocket
	err = discord.Open()
	if err != nil {
		fmt.Println(err)
		return
	}

	//// Connect to voice channel.
	//// NOTE: Setting mute to true, deaf to false.
	//dgv, err := discord.ChannelVoiceJoin(*GuildID, *ChannelID, true, false)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}

	c, err = oto.NewContext(*sampleRate, *channelNum, *bitDepthInBytes, 4096)
	if err != nil {
		panic(err.Error())
	}
	player = c.NewPlayer()

	// Register the messageCreate func as a callback for MessageCreate events.
	discord.AddHandler(messageCreate)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Close()

	err = player.Close()
	if err != nil {
		panic(err.Error())
	}
	c.Close()

	return
}

var (
	GuildID   = flag.String("g", "430111136822722590", "Guild ID")
	ChannelID = flag.String("c", "643492477822828554", "Channel ID")
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Only process valid commands
	if len(m.Content) > 1 && m.Content[:1] == "!" {
		log.Print(m.Content)
		method := strings.Split(m.Content, " ")[0][1:]

		if method == "j" || method == "join" {
			if len(s.VoiceConnections) == 0 || s.VoiceConnections[*GuildID].ChannelID != *ChannelID {
				// Connect to voice channel.
				// NOTE: Setting mute to true, deaf to false.
				dgv, err := s.ChannelVoiceJoin(*GuildID, *ChannelID, true, false)
				if err != nil {
					fmt.Println(err)
					return
				}

				// Starts listen
				listen(dgv)

				// Close connections
				//dgv.Close()
			}
		}

		if method == "l" || method == "leave" {
			// Connect to voice channel.
			// NOTE: Setting mute to true, deaf to false.
			s.VoiceConnections[*GuildID].Close()
			_, err := s.ChannelVoiceJoin(*GuildID, "", true, false)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

var (
	sampleRate      = flag.Int("samplerate", 44100, "sample rate")
	channelNum      = flag.Int("channelnum", 2, "number of channel")
	bitDepthInBytes = flag.Int("bitdepthinbytes", 1, "bit depth in bytes")
)

func listen(v *discordgo.VoiceConnection) {

	recv := make(chan *discordgo.Packet, 2)
	go dgvoice.ReceivePCM(v, recv)

	for {
		// todo: will this loop finish when the bot leaves the channel?
		p, ok := <-recv
		if !ok {
			print("Not okay")
			return
		}

		bytes := make([]byte, len(p.PCM))
		for index, _ := range bytes {
			bytes[index] = uint8(p.PCM[index]>>8) + 128
		}

		_, err := player.Write(bytes)
		if err != nil {
			print(err.Error())
			return
		}
	}
}
