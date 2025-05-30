# AntiDcGenAI


Discord recently added a feature to edit/"reimagine" images using generative AI tools, which cannot be disabled by server owners/admins.  
Therefore, I made this: A basic bot to automatically delete the messages and, optionally, warn/timeout/kick/ban people using them.  
Written in Go to become more familiar with it.

## Option 1:

Use the hosted bot, [**add here**](https://discord.com/oauth2/authorize?client_id=1366214068532285541&permissions=1374389544966&integration_type=0&scope=bot).
You 
This version has warn + kick + 1h timeout (but not ban) active by default.
**Set with care after adding the bot**

## Option 2:
Host it yourself. It's easy to do, runs on pretty much anything with an internet connection.

---

## Prerequisites

- A Discord bot token. Create a new bot/app in the Discord developer portal for that.
- Podman or Docker (optional but recommended)
- To run without a container, ideally a tool like `screen` or `tmux`

---

## Setup
Run the bot from the directory you want the config in (or set `CONFIG_PATH` to change the location)
After the first start, the config (`config\config.json`) will be generated.  
Fill in the bot token and just run it again.


#### Directly (without a container):
1. Download [the latest release](https://github.com/itsTyrion/AntiDiscordGenAI/releases/latest). If unsure, you probably want amd64 or arm64 for desktop/laptop/server and arm64/armv7/armv6 for e.g. a Raspberry Pi. (darwin = MacOS)
2. Just run it. On a server, use `screen` to keep it running after disconnecting

#### With Podman:
1. git clone the repo (or download as zip and unzip it)
2. 
```
podman build -t anti-dc-genai:latest .

podman run -d \
  --name AntiDcGenAI \
  --memory=32m --cpus="0.5" \
  -v "$(pwd)/config:/app/config:Z" \
  --restart=unless-stopped \
  anti-dc-genai:latest
```
#### With Docker:
1. git clone the repository (or download as zip and unzip it)
2. create the config directory `mkdir config` to avoid docker permission problems
3. 
```
docker build -t anti-dc-genai:latest .

docker run -d \
  --name AntiDcGenAI \
  --memory=32m --cpus="0.5" \
  -v "$(pwd)/config:/app/config" \
  --restart=unless-stopped \
  anti-dc-genai:latest
```

## Config

Replace `"YOUR_TOKEN_HERE"` with your actual Discord bot token.  
Change config options to your liking or just configure via the command.  
If you come across more of those bots/apps, add their user ID to the list and the bot will handle them with a warn (+deletion if configured) as well, but no kick/ban. You're welcome to open an issue so I can properly handle them.

---

## Notes

- The bot reads its configuration from a configurable folder (`./config/config,json` w/ default commands) with podman/dockerdocker, and $CONFIG_PATH/config.json (./config.json if unset) outside of it.
- Make sure your're running the Docker commands from the path you want the config folder in or adjust accordingly. 
- To stop the container:

```bash
podman stop AntiDcGenAI
podman rm AntiDcGenAI
```
or
```bash
docker stop AntiDcGenAI
docker rm AntiDcGenAI
```

---

## License

This project is licensed under MIT. You can find the full license text in the `LICENSE` file.
