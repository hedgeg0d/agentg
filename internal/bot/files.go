package bot

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) downloadCommand(msg *tgbotapi.Message) {
	chat := msg.Chat.ID
	path := strings.TrimSpace(msg.CommandArguments())
	if path == "" {
		b.send(chat, "Usage: `/download <path>`")
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		b.send(chat, "⚠️ "+err.Error())
		return
	}
	if info.IsDir() {
		b.send(chat, "⚠️ Path is a directory.")
		return
	}
	doc := tgbotapi.NewDocument(chat, tgbotapi.FilePath(path))
	doc.Caption = path
	if _, err := b.api.Send(doc); err != nil {
		b.send(chat, "⚠️ "+err.Error())
	}
}

func (b *Bot) handleUpload(msg *tgbotapi.Message) {
	fileID, name := uploadTarget(msg)
	if fileID == "" {
		return
	}
	dest := b.resolveDest(msg.Caption, name)
	if err := b.saveTelegramFile(fileID, dest); err != nil {
		b.send(msg.Chat.ID, "⚠️ "+err.Error())
		return
	}
	b.send(msg.Chat.ID, "✅ Saved to `"+dest+"`")
}

func uploadTarget(msg *tgbotapi.Message) (string, string) {
	switch {
	case msg.Document != nil:
		return msg.Document.FileID, msg.Document.FileName
	case len(msg.Photo) > 0:
		p := msg.Photo[len(msg.Photo)-1]
		return p.FileID, p.FileUniqueID + ".jpg"
	case msg.Video != nil:
		return msg.Video.FileID, msg.Video.FileName
	case msg.Audio != nil:
		return msg.Audio.FileID, msg.Audio.FileName
	default:
		return "", ""
	}
}

func (b *Bot) resolveDest(caption, name string) string {
	if name == "" {
		name = "upload"
	}
	caption = strings.TrimSpace(caption)
	if caption == "" {
		return filepath.Join(b.cfg.UploadDir, name)
	}
	if info, err := os.Stat(caption); err == nil && info.IsDir() {
		return filepath.Join(caption, name)
	}
	if strings.HasSuffix(caption, "/") {
		return filepath.Join(caption, name)
	}
	return caption
}

func (b *Bot) saveTelegramFile(fileID, dest string) error {
	f, err := b.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return err
	}
	resp, err := http.Get(f.Link(b.api.Token))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if dir := filepath.Dir(dest); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}
