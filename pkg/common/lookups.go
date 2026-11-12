package common

var mrState2EmojiMap = map[string]string{
	"merged": "✅",
	"opened": "🟢",
	"closed": "❌",
}

func LookupMRState2Emoji(state string) string {
	e, ok := mrState2EmojiMap[state]
	if ok {
		return e
	}
	return "❔"
}
