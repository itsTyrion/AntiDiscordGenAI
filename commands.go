package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type CommandHandler struct {
	Session *discordgo.Session
	Config  *Config
}

var GuildSettingsCommand = discordgo.ApplicationCommand{
	Name:        "settings",
	Description: "Configure anti-discord settings for this guild.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "view",
			Description: "View current settings for this guild.",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
		},
		{
			Name:        "set",
			Description: "Set a specific setting for this guild.",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "option",
					Description: "The setting to configure",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Delete Messages", Value: "delete"},
						{Name: "Warn Users", Value: "warn"},
						{Name: "Kick Users", Value: "kick"},
						{Name: "Ban Users", Value: "ban"},
						{Name: "Timeout Duration (seconds)", Value: "timeout"},
						{Name: "Warn Message, \\n for newlines", Value: "warn_message"},
					},
				},
				{
					Name:        "value",
					Description: "New value (true/false, seconds for timeout, message string)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		{
			Name:        "bots",
			Description: "Manage blacklisted bots",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "action",
					Description: "Action to perform",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Add Bot", Value: "add"},
						{Name: "Remove Bot", Value: "remove"},
						{Name: "List Bots", Value: "list"},
					},
				},
				{
					Name:        "bot_id",
					Description: "Bot user to add/remove",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    false,
				},
			},
		},
	},
}

func RegisterCommands(s *discordgo.Session, c *Config, logger *zap.SugaredLogger) {
	handler := &CommandHandler{Session: s, Config: c}

	if _, err := s.ApplicationCommandCreate(s.State.User.ID, "", &GuildSettingsCommand); err != nil {
		logger.Errorf("Error creating command: %v\n", err)
	}
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		data := i.ApplicationCommandData()

		switch data.Name {
		case "settings":
			if i.GuildID == "" {
				respondEphemeral(s, i, "This command can only be used in a guild.")
				return
			}
			options := data.Options
			if len(options) == 0 {
				return
			}

			switch options[0].Name {
			case "view":
				handler.handleViewSettings(s, i)
			case "set":
				handler.handleSetSettings(s, i)
			case "bots":
				handler.handleBotsCommand(s, i)
			}
		default:
			logger.Errorf("Unknown command received: %s\n", data.Name)
		}
	})

	logger.Info("Guild settings commands registered and handler added.")

}
func (h *CommandHandler) handleViewSettings(s *discordgo.Session, i *discordgo.InteractionCreate) {

	settings := h.Config.ForGuildID(i.GuildID)
	response := fmt.Sprintf(
		"**Current Settings:**\n"+
			"Delete Messages: `%t`\n"+
			"Warn Users: `%t`\n"+
			"Kick Users: `%t`\n"+
			"Ban Users: `%t`\n"+
			"Timeout Duration (seconds): `%d`\n"+
			"Warn Message: ```%s```",
		settings.Delete, settings.Warn, settings.Kick, settings.Ban, settings.Timeout, settings.WarnMessage,
	)

	respondEphemeral(s, i, response)
}

func (h *CommandHandler) handleSetSettings(s *discordgo.Session, i *discordgo.InteractionCreate) {

	opts := i.ApplicationCommandData().Options[0].Options
	var setting string
	var rawValue string
	var response string

	for _, opt := range opts {
		switch opt.Name {
		case "option":
			setting = opt.StringValue()
		case "value":
			rawValue = opt.StringValue()
		}
	}

	settings := h.Config.ForGuildID(i.GuildID)
	newSettings := settings.Copy()
	parseBool := func(value string) (boolValue bool, err error) {
		boolValue, err = strconv.ParseBool(value)
		if err != nil {
			respondEphemeral(s, i, "Invalid value for delete. Please use `true` or `false`.")
		}
		return
	}

	switch setting {
	case "delete":
		if boolValue, err := parseBool(rawValue); err == nil {
			newSettings.Delete = boolValue
			response = fmt.Sprintf("Set `Delete Messages` to `%t`.", boolValue)
		}
	case "warn":
		if boolValue, err := parseBool(rawValue); err == nil {
			newSettings.Warn = boolValue
			response = fmt.Sprintf("Set `Warn Users` to `%t`.", boolValue)
		}
	case "kick":
		if boolValue, err := parseBool(rawValue); err == nil {
			newSettings.Kick = boolValue
			response = fmt.Sprintf("Set `Kick Users` to `%t`.", boolValue)
		}
	case "ban":
		if boolValue, err := parseBool(rawValue); err == nil {
			newSettings.Ban = boolValue
			response = fmt.Sprintf("Set `Ban Users` to `%t`.", boolValue)
		}
	case "timeout":
		intValue, err := strconv.Atoi(rawValue)
		if err != nil || intValue < 0 {
			respondEphemeral(s, i, "Invalid value for timeout. Please use a non-negative number or `0` to disable.")
			return
		}
		newSettings.Timeout = intValue
		if intValue == 0 {
			response = "Set Timeout to `0` seconds (disabled)."
		} else {
			response = fmt.Sprintf("Set Timeout to `%d` seconds.", intValue)
		}
	case "warn_message":
		newSettings.WarnMessage = strings.ReplaceAll(rawValue, "\\n", "\n")
		response = fmt.Sprintf("Set `Warn Message` to: ```%s```", newSettings.WarnMessage)
	default:
		respondEphemeral(s, i, "Unknown setting option")
		return
	}

	h.Config.UpdateGuildSettings(i.GuildID, newSettings)
	respondEphemeral(s, i, response)
}

func (h *CommandHandler) handleBotsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	var action string
	var botUser *discordgo.User

	for _, opt := range i.ApplicationCommandData().Options[0].Options {
		switch opt.Name {
		case "action":
			action = opt.StringValue()
		case "bot_id":
			botUser = opt.UserValue(s)
		}
	}

	if botUser == nil && action != "list" {
		respondEphemeral(s, i, "Please specify a bot user")
		return
	}

	settings := h.Config.ForGuildID(i.GuildID)
	newSettings := settings.Copy()
	response := ""

	switch action {
	case "add":
		if !botUser.Bot {
			respondEphemeral(s, i, "Specified user is not a bot")
			return
		}
		if slices.Contains(newSettings.Bots, botUser.ID) {
			respondEphemeral(s, i, "Bot is already blacklisted")
			return
		}

		newSettings.Bots = append(newSettings.Bots, botUser.ID)
		h.Config.UpdateGuildSettings(i.GuildID, newSettings)
		response = fmt.Sprintf("Added bot: <@%s>", botUser.ID)

	case "remove":
		if idx := slices.Index(newSettings.Bots, botUser.ID); idx != -1 {
			newSettings.Bots = slices.Delete(newSettings.Bots, idx, idx+1)
			h.Config.UpdateGuildSettings(i.GuildID, newSettings)
			response = fmt.Sprintf("Removed bot: <@%s>", botUser.ID)
		} else {
			respondEphemeral(s, i, "Bot not found in blacklisted list")
			return
		}

	case "list":
		if len(settings.Bots) == 0 {
			respondEphemeral(s, i, "No blacklisted bots")
			return
		}
		response = "**Blacklisted Bots:**\n"
		for _, botID := range settings.Bots {
			response += fmt.Sprintf("- <@%s>\n", botID)
		}

	default:
		respondEphemeral(s, i, "Invalid action")
		return
	}

	respondEphemeral(s, i, response)
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content, Flags: discordgo.MessageFlagsEphemeral},
	})
}
