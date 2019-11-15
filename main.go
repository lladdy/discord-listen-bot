package main

import (
	"flag"
	"fmt"
	"github.com/hajimehoshi/oto"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

func main() {

	var (
		GuildID   = flag.String("g", "644694725202542603", "Guild ID")
		ChannelID = flag.String("c", "644694725202542607", "Channel ID")
		err       error
	)
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

	// Connect to voice channel.
	// NOTE: Setting mute to false, deaf to true.
	dgv, err := discord.ChannelVoiceJoin(*GuildID, *ChannelID, false, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Starts echo
	echo(dgv)

	// Close connections
	dgv.Close()
	discord.Close()

	return
}

var (
	sampleRate      = flag.Int("samplerate", 44100, "sample rate")
	channelNum      = flag.Int("channelnum", 2, "number of channel")
	bitDepthInBytes = flag.Int("bitdepthinbytes", 1, "bit depth in bytes")
)

func echo(v *discordgo.VoiceConnection) {

	recv := make(chan *discordgo.Packet, 2)
	go dgvoice.ReceivePCM(v, recv)

	send := make(chan []int16, 2)
	go dgvoice.SendPCM(v, send)

	v.Speaking(true)
	defer v.Speaking(false)

	c, err := oto.NewContext(*sampleRate, *channelNum, *bitDepthInBytes, 4096)
	if err != nil {
		panic(err.Error())
	}
	player := c.NewPlayer()
	defer player.Close()

	for {

		p, ok := <-recv
		if !ok {
			return
		}

		bytes := make([]byte, len(p.PCM))
		for index, _ := range bytes {
			bytes[index] = uint8(p.PCM[index]>>8) + 128
		}

		_, err = player.Write(bytes)
		if err != nil {
			panic(err.Error())
		}
	}
}
