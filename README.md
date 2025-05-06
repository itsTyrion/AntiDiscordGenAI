# AntiDcGenAI


Discord recently added a feature to edit/"reimagine" images using generative AI tools, which cannot be disabled by server owners/admins.  
Therefore, I made this: A basic bot to at  automatically delete the messages and, optionally, warn/kick/ban people using them.  
Written in Go to become more familiar with it.

## Option 1:

Use the hosted bot, [**add here**](https://discord.com/oauth2/authorize?client_id=1366214068532285541&permissions=1374389544966&integration_type=0&scope=bot).
This version always warns, and deletes/kicks/*bans* depending on the permissions.  
**Set with care.**

## Option 2:
Host it yourself. It's easy to do, runs on pretty much anything with an internet connection and gives you a bit more config control

---

## Prerequisites

- A Discord bot token. Create a new bot/app in the Discord developer portal for that.
- Podman or Docker (optional but recommended)
- To run without a container, ideally a tool like `screen` or `tmux`

---

## Setup
Run the bot from the directory you want the config in (or set `CONFIG_PATH` to change the location)
After the first start, the config (`config.json`) will be generated.  
Fill in the bot token and just run it again.


#### Directly (without a container):
1. Download [the latest release](actions). If unsure, you probably want amd64 or arm64 for desktop/laptop/server and arm64/armv7/armv6 for e.g. a Raspberry Pi. (darwin = MacOS)
2. Just run it. On a server, use `screen` to keep it running after disconnecting

#### With Podman:
1. git glone the repo (or download as zip and unzip it)
2. 
```
podman build -t anti-dc-genai:latest .

podman run -d \
  --name AntiDcGenAI \
  --memory=32m --cpus="0.5" \
  -v "$(pwd)/config.json:/app/config.json:Z" \
  anti-dc-genai:latest
```
#### With Docker:
1. git clone the repository (or download as zip and unzip it)
2. 
```
podman build -t anti-dc-genai:latest .

docker run -d \
  --name AntiDcGenAI \
  --memory=32m --cpus="0.5" \
  -v "$(pwd)/config.json:/app/config.json" \
  anti-dc-genai:latest
```

## Config

Replace `"YOUR_TOKEN_HERE"` with your actual Discord bot token.  
Change `delete`/`warn`/`kick`/`ban` to your liking (`true`/`false`).  
If you come across more of those bots/apps, add their user ID to the list and the bot will handle them with a warn (+deletion if configured) as well, but no kick/ban.

---

## Notes

- The bot reads its configuration from `/app/config.json` inside the container, and $CONFIG_PATH/config.json (./config.json if unset) outside of it.
- Make sure your `config.json` is in the same directory as the Docker commands or adjust paths accordingly.  
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
