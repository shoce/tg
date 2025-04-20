package tg

import (
	"testing"
)

// https://pkg.go.dev/testing

func TestTg(t *testing.T) {

	msg := ""
	msg += Esc("special chars: \\*_[]()~`>#+-=|{}.!") + NL +
		Bold("bold") + NL +
		Italic("italic") + NL +
		Underline("underline") + NL +
		BoldUnderline("bold underline") + NL +
		ItalicUnderline("italic underline") + NL +
		Spoiler("spoiler") + NL +
		Code("code") + NL +
		Link("bots api link", "https://core.telegram.org/bots/api") + NL +
		Pre("  pre * "+NL+"  * - * "+NL+"  * formatted") + NL +
		Quote("normal"+NL+"quote"+NL+"block"+NL+"shows"+NL+"all"+NL+"lines") + NL +
		ExpQuote("expandable"+NL+"quote"+NL+"block"+NL+"hides"+NL+"lines"+NL+"until"+NL+"expanded")

	log("message:" + NL + msg + NL + ":")

	ApiToken = ""

	if r, err := SendMessage(SendMessageRequest{
		ChatId: "",
		Text:   msg,

		LinkPreviewOptions: LinkPreviewOptions{IsDisabled: true},
	}); err != nil {
		t.Error(err)
	} else {
		log("SendMessage result message id==%v", r.Id)
	}

}
