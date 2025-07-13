package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	IsEnableMail bool
	MailTo       string
	MailSubject  string
	MailHost     string
	MailPort     string
	MailUsername string
	MailPassword string

	IsEnableTelegram bool
	TelegramToken    string
	TelegramChatID   string

	IsEnableDiscord   bool
	DiscordWebhookURL string

	IsEnableLine     bool
	LineChannelToken string

	LineUserID      string
	LogFile         string
	ThrottleSeconds int
}

func loadEnv() Config {
	throttleSeconds, _ := strconv.Atoi(os.Getenv("THROTTLE_SECONDS"))

	return Config{
		IsEnableMail: os.Getenv("ENABLE_MAIL") == "true",
		MailTo:       os.Getenv("MAIL_TO"),
		MailSubject:  os.Getenv("MAIL_SUBJECT"),
		MailHost:     os.Getenv("MAIL_HOST"),
		MailPort:     os.Getenv("MAIL_PORT"),
		MailUsername: os.Getenv("MAIL_USERNAME"),
		MailPassword: os.Getenv("MAIL_PASSWORD"),

		IsEnableTelegram: os.Getenv("ENABLE_TELEGRAM") == "true",
		TelegramToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),

		IsEnableDiscord:   os.Getenv("ENABLE_DISCORD") == "true",
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),

		IsEnableLine:     os.Getenv("ENABLE_LINE") == "true",
		LineChannelToken: os.Getenv("LINE_CHANNEL_TOKEN"),
		LineUserID:       os.Getenv("LINE_USER_ID"),

		LogFile:         os.Getenv("LOG_FILE"),
		ThrottleSeconds: throttleSeconds,
	}
}

func sendEmail(cfg Config, msg string) {
	auth := smtp.PlainAuth("", cfg.MailUsername, cfg.MailPassword, cfg.MailHost)
	to := []string{cfg.MailTo}
	msgBody := fmt.Appendf(nil, "Subject: %s\r\n\r\n%s", cfg.MailSubject, msg)
	err := smtp.SendMail(cfg.MailHost+":"+cfg.MailPort, auth, cfg.MailUsername, to, msgBody)
	if err != nil {
		log.Printf("Email send failed: %v", err)
	}
}

func sendTelegram(cfg Config, msg string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.TelegramToken)
  _, err := http.PostForm(url, map[string][]string{
		"chat_id": {cfg.TelegramChatID},
		"text":    {msg},
	})
	if err != nil {
		log.Printf("Telegram send failed: %v", err)
	}
}

func sendDiscord(cfg Config, msg string) {
	data := map[string]string{"content": msg}
	jsonBody, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", cfg.DiscordWebhookURL, strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
  _, err := client.Do(req)
	if err != nil {
		log.Printf("Discord send failed: %v", err)
	}
}

func sendLine(cfg Config, msg string) {
	if cfg.LineChannelToken == "" || cfg.LineUserID == "" {
		return
	}
	body := map[string]interface{}{
		"to": cfg.LineUserID,
		"messages": []map[string]string{{
			"type": "text",
			"text": msg,
		}},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "https://api.line.me/v2/bot/message/push", strings.NewReader(string(jsonBody)))
	req.Header.Set("Authorization", "Bearer "+cfg.LineChannelToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("LINE Messaging API failed: %v", err)
		return
	}
	defer resp.Body.Close()
}

func sendAlert(cfg Config, message string) {
	fmt.Printf("[%s] ALERT: %s\n", time.Now().Format(time.RFC3339), message)
	if cfg.IsEnableMail {
		go sendEmail(cfg, message)
	}
	if cfg.IsEnableTelegram {
		go sendTelegram(cfg, message)
	}
	if cfg.IsEnableDiscord {
		go sendDiscord(cfg, message)
	}
	if cfg.IsEnableLine {
		go sendLine(cfg, message)
	}
}

func main() {
	cfg := loadEnv()
	lastAlertFile := "/tmp/last_5xx_alert"
	lastAlert := int64(0)

	if b, err := os.ReadFile(lastAlertFile); err == nil {
		lastAlert, _ = strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	}

	f, err := os.Open(cfg.LogFile)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer f.Close()

	f.Seek(0, io.SeekEnd)
	r := bufio.NewReader(f)
	re := regexp.MustCompile(` 5[0-9]{2} `)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if re.MatchString(line) {
			now := time.Now().Unix()
			if now-lastAlert >= int64(cfg.ThrottleSeconds) {
				sendAlert(cfg, strings.TrimSpace(line))
				os.WriteFile(lastAlertFile, fmt.Appendf(nil, "%d", now), 0644)
				lastAlert = now
			}
		}
	}
}
