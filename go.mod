module DiscordBot

go 1.23

toolchain go1.23.5

require (
	QCheckWE v0.0.0
	github.com/PuerkitoBio/goquery v1.10.1
	github.com/bwmarrin/discordgo v0.28.1
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
)

replace QCheckWE => ./QCheckWE

require (
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/go-resty/resty/v2 v2.16.4 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)
