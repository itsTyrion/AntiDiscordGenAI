package main

import (
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func main() {
	logger := SetupLogging()

	config, err := LoadConfig()
	if err != nil {
		logger.Errorf("Failed to load config: %v", err)
		time.Sleep(5 * time.Second)
		return
	}

	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		logger.Fatalf("Failed to create Discord dg: %v", err) // Fatalf calls exit
	}
	dg.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentMessageContent | discordgo.IntentGuildMembers

	dg.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) { handleMsg(dg, m.Message, logger, config) })
	dg.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageUpdate) { handleMsg(dg, m.Message, logger, config) })

	if err = dg.Open(); err != nil {
		logger.Fatalf("Failed to connect to Discord: %v", err)
		return
	}
	logger.Info("Bot is now running. Press CTRL-C to exit.")
	defer dg.Close()

	select {}
}

func handleMsg(dg *discordgo.Session, m *discordgo.Message, logger *zap.SugaredLogger, config *Config) {
	state := dg.State
	if m.Author.ID == state.User.ID || !m.Author.Bot || !slices.Contains(config.Bots, m.Author.ID) {
		return
	}

	channelPerms, err := state.UserChannelPermissions(state.User.ID, m.ChannelID)
	if err != nil {
		logger.Errorf("Failed to get permissions: %v", err)
		return
	}
	lacksPermission := func(perm int64) bool { return channelPerms&perm == 0 }
	reason := "Use of Discord's 'AI' tools is against our rules"

	if config.Delete {
		if lacksPermission(discordgo.PermissionManageMessages) {
			logger.Infof("Bot lacks manage messages permission, skipping deletion")
		} else if err := dg.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
			logger.Errorf("Failed to delete message: %v", err)
		}
	}
	switch m.Author.ID {
	case "1104973139257081909": // Viggle
	case "1153984868804468756": // DomoAI
	case "1288638725869535283": // Glifbot
	case "1090660574196674713": // InsightFaceSwap
		if len(m.Mentions) == 0 {
			return
		}
		user := m.Mentions[0]

		if config.Warn {
			if _, err := dg.ChannelMessageSend(m.ChannelID, "<@"+user.ID+">, please adhere to the rules!"); err != nil {
				logger.Errorf("Failed to send warning: %v", err)
			}
		}
		if config.Kick {
			if lacksPermission(discordgo.PermissionKickMembers) {
				logger.Infof("Bot lacks kick members permission, skipping kick")
			} else {
				if err := dg.GuildMemberDeleteWithReason(m.GuildID, user.ID, reason); err != nil {
					logger.Errorf("Failed to kick member: %v", err)
				}
				logger.Infof("%s (%s) was kicked: %s", user.Username, user.ID, reason)
				if _, err := dg.ChannelMessageSend(m.ChannelID, "<@"+user.ID+"> was kicked: "+reason); err != nil {
					logger.Errorf("Failed to send kick message: %v", err)
				}
			}
		}
		if config.Ban {
			if lacksPermission(discordgo.PermissionBanMembers) {
				logger.Infof("Bot lacks ban members permission, skipping ban")
			} else {
				if err := dg.GuildBanCreateWithReason(m.GuildID, user.ID, reason, 0); err != nil {
					logger.Errorf("Failed to ban member: %v", err)
				}
				logger.Infof("%s (%s) was banned: %s", user.Username, user.ID, reason)
				if _, err := dg.ChannelMessageSend(m.ChannelID, "<@"+user.ID+"> was banned: "+reason); err != nil {
					logger.Errorf("Failed to send ban message: %v", err)
				}
			}
		}
	}
}
