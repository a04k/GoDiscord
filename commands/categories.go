package commands

var CommandCategories = map[string][]string{
	"General":      {"help", "commandlist", "usd", "btc", "remindme"},
	"Economy":      {"balance", "work", "transfer", "flip", "setdailyrole", "removedailyrole", "listdailyroles"},
	"EPL":          {"epltable", "nextmatch"},
	"F1":           {"f1", "f1results", "f1standings", "f1wdc", "f1wcc", "qualiresults", "nextf1session", "f1sub"},
	"Fpl":          {"fplstandings", "setfplleague"},
	"Moderation":   {"kick", "mute", "unmute", "voicemute", "vunmute", "ban", "unban"},
	"Admin":        {"setadmin", "add", "take", "disable", "enable"},
	"Roles":        {"createrole", "setrole", "removerole", "inrole", "roleinfo"},
}
