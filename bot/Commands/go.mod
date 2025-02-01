module commands

go 1.20

require (
	WeQuota v0.0.0
	github.com/bwmarrin/discordgo v0.28.1
	utils v0.0.0
)

replace WeQuota => ../scripts/WeQuota

replace utils => ../Utils

require (
	github.com/go-resty/resty/v2 v2.16.4 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)
