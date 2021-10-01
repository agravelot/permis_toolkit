package discord

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

var channelID string

func Start(token string) (*discordgo.Session, error) {
	dg, err := discordgo.New("Bot " + token)

	if err != nil {
		return nil, err
	}

	err = dg.Open()
	if err != nil {
		return nil, err
	}

	for _, g := range dg.State.Guilds {
		channels, err := dg.GuildChannels(g.ID)
		if err != nil {
			return nil, err
		}
		for _, c := range channels {
			// TODO Make it configurable
			if c.Name == "permis" {
				channelID = c.ID
			}
		}
	}

	if channelID == "" {
		return nil, errors.New("unable to get channel id")
	}

	return dg, nil
}

func Notify(dg *discordgo.Session, message string) error {
	// TODO use ChannelMessageSendComplex to not include embed og image
	_, err := dg.ChannelMessageSend(channelID, message)
	if err != nil {
		return err
	}

	return nil
}
