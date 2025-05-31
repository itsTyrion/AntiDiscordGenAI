package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

func main() {
	logger := SetupLogging()

	config, err := LoadConfig()
	if err != nil {
		logger.Errorf("Failed to load config: %v", err)
		time.Sleep(5 * time.Second)
		return
	}
	InitConfigSaver(config)

	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		logger.Fatalf("Failed to create Discord dg: %v", err) // Fatalf calls exit
	}
	dg.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentGuilds

	dg.AddHandler(func(_ *discordgo.Session, m *discordgo.Ready) {
		logger.Info("Ready")
		RegisterCommands(dg, config, logger)
	})
	dg.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		handleMsg(dg, m.Message, logger, config)
	})

	if err = dg.Open(); err != nil {
		logger.Fatalf("Failed to connect to Discord: %v", err)
		return
	}
	logger.Info("Bot is now running. Press CTRL-C to exit.")
	defer dg.Close()

	select {}
}
