package discord

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

var session *discordgo.Session
var channelID string

func Start(token string) error {
	dg, err := discordgo.New("Bot " + token)

	// TODO Do not use global ?
	session = dg

	if err != nil {
		return err
	}

	err = session.Open()
	if err != nil {
		return err

	}

	for _, g := range session.State.Guilds {
		channels, err := session.GuildChannels(g.ID)
		if err != nil {
			return err
		}
		for _, c := range channels {
			if c.Name == "permis" {
				channelID = c.ID
			}
		}
	}

	if channelID == "" {
		return errors.New("unable to get channel id")
	}

	return nil
}

func Close() {
	session.Close()
}

func Notify(message string) error {
	// TODO use ChannelMessageSendComplex to not include embed og image
	_, err := session.ChannelMessageSend(channelID, message)
	if err != nil {
		return err
	}

	return nil
}
