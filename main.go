package main

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	botToken := os.Getenv("TELEGRAM_API_KEY")

	// Telegram Bot API
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// Telegram Update
	updates, err := bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	if err != nil {
		log.Panic(err)
	}
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() { // 커맨드 메시지 감지
			switch update.Message.Command() {
			case "trigger":
				// PagerDuty 알림을 전송하는 함수 호출
				sendPagerDutyAlert()
				replyMessage := tgbotapi.NewMessage(update.Message.Chat.ID, "PagerDuty alert triggered.")
				bot.Send(replyMessage)
			}
		}
	}
}

func sendPagerDutyAlert() {
	// PagerDuty Generic Events API 엔드포인트 및 토큰 설정
	pagerDutyEndpoint := "https://events.pagerduty.com/v2/enqueue"
	pagerDutyIntegrationKey := os.Getenv("PAGERDUTY_INTEGRATION_KEY")
	pagerDutyToken := os.Getenv("PAGERDUTY_TOKEN")

	requestBody := []byte(`{
		"event_action": "trigger",
		"routing_key": "` + pagerDutyIntegrationKey + `",
		"payload": {
			"summary": "Telegram Oncall Triggered",
			"severity": "critical",
			"source": "Telegram Bot"
		}
	}`)

	pagerDutyURL, err := url.Parse(pagerDutyEndpoint)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest("POST", pagerDutyURL.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pagerDutyToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()

	// HTTP Response
	if resp.StatusCode != http.StatusAccepted {
		log.Printf("HTTP request failed. Status code: %d", resp.StatusCode)
	}
}
