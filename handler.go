package main

import (
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func handleMsg(dg *discordgo.Session, m *discordgo.Message, logger *zap.SugaredLogger, c *Config, b bool) {
	config := c.ForGuildID(m.GuildID)
	interaction := m.Interaction

	if interaction == nil || !m.Author.Bot || !slices.Contains(config.Bots, m.Author.ID) {
		return
	}

	logger.Infof("Processing message from %s: %s", m.Author.Username, m.Content)

	state := dg.State
	channelPerms, err := state.UserChannelPermissions(state.User.ID, m.ChannelID)
	if err != nil {
		logger.Errorf("Failed to get permissions for channel %s: %v", m.ChannelID, err)
		return
	}
	lacksPermission := func(perm int64) bool { return channelPerms&perm == 0 }
	reason := config.WarnMessage
	logReason := discordgo.WithAuditLogReason(reason)
	user := interaction.User

	if config.Delete {
		if lacksPermission(discordgo.PermissionManageMessages) {
			logger.Infof("Bot lacks manage messages permission, skipping deletion")
		} else if err := dg.ChannelMessageDelete(m.ChannelID, m.ID, logReason); err != nil {
			logger.Errorf("Failed to delete message: %v", err)
		} else {
			logger.Infof("Deleted message from %s", m.Author.Username)
		}
	}

	if config.Warn {
		if _, err := dg.ChannelMessageSend(m.ChannelID, "<@"+user.ID+"> "+reason); err != nil {
			logger.Errorf("Failed to send warning: %v", err)
		} else {
			logger.Infof("Warning issued to user: %s (mentioned by bot %s)", user.ID, m.Author.Username)
		}
	}

	if config.Timeout > 0 {
		if lacksPermission(discordgo.PermissionModerateMembers) {
			logger.Infof("Bot lacks moderate members permission, skipping timeout")
		} else {
			timeoutDuration := time.Duration(config.Timeout) * time.Second
			until := time.Now().UTC().Add(timeoutDuration)
			if err := dg.GuildMemberTimeout(m.GuildID, user.ID, &until, logReason); err != nil {
				logger.Errorf("Failed to timeout member: %v", err)
			} else {
				logger.Infof("Timed out user %s for %d seconds", user.ID, config.Timeout)
			}
		}
	}

	if config.Kick {
		if lacksPermission(discordgo.PermissionKickMembers) {
			logger.Infof("Bot lacks kick members permission, skipping kick")
		} else {
			if err := dg.GuildMemberDeleteWithReason(m.GuildID, user.ID, reason, logReason); err != nil {
				logger.Errorf("Failed to kick member: %v", err)
			} else {
				logger.Infof("%s (%s) was kicked: %s", user.Username, user.ID, reason)
				if _, err := dg.ChannelMessageSend(m.ChannelID, "<@"+user.ID+"> was kicked: "+reason); err != nil {
					logger.Errorf("Failed to send kick message: %v", err)
				}
			}
		}
	}

	if config.Ban {
		if lacksPermission(discordgo.PermissionBanMembers) {
			logger.Infof("Bot lacks ban members permission, skipping ban")
		} else {
			if err := dg.GuildBanCreateWithReason(m.GuildID, user.ID, reason, 0, logReason); err == nil {
				logger.Errorf("Failed to ban member: %v", err)
			} else {
				logger.Infof("%s (%s) was banned: %s", user.Username, user.ID, reason)
				if _, err := dg.ChannelMessageSend(m.ChannelID, "<@"+user.ID+"> was banned: "+reason); err != nil {
					logger.Errorf("Failed to send ban message: %v", err)
				}
			}
		}
	}
}
