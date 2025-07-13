#!/bin/bash
set -e

# Load environment
source .env

LAST_ALERT_FILE="/tmp/last_5xx_alert"
touch "$LAST_ALERT_FILE"

send_alert() {
  local message="$1"
  echo "[$(date)] ALERT: $message"

  # Email
  if [ "$ENABLE_MAIL" = "true" ]; then
    {
      echo "Subject: $MAIL_SUBJECT"
      echo "To: $MAIL_TO"
      echo
      echo "$message"
    } | msmtp "$MAIL_TO"
  fi

  # Telegram
  if [ "$ENABLE_TELEGRAM" = "true" ]; then
    curl -s -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/sendMessage" \
      -d chat_id="$TELEGRAM_CHAT_ID" \
      -d text="$message" >/dev/null
  fi

  ESCAPED_MSG=$(printf '%s' "$message" | sed 's/"/\\"/g')
  # Discord
  if [ "$ENABLE_DISCORD" = "true" ]; then
    curl -s -X POST "$DISCORD_WEBHOOK_URL" \
      -H "Content-Type: application/json" \
      -d "{\"content\": \"$ESCAPED_MSG\"}" >/dev/null
  fi

  # LINE
  if [ "$ENABLE_LINE" = "true" ]; then
    curl -s -X POST "https://api.line.me/v2/bot/message/push" \
      -H "Authorization: Bearer $LINE_CHANNEL_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"to\":\"$LINE_USER_ID\",\"messages\":[{\"type\":\"text\",\"text\":\"$ESCAPED_MSG\"}]}" >/dev/null
  fi
}

# Monitor log file
tail -F "$LOG_FILE" | while read -r line; do
  if echo "$line" | grep -qE " 5[0-9]{2} "; then
    NOW=$(date +%s)
    LAST=$(cat "$LAST_ALERT_FILE")
    if ((NOW - LAST >= THROTTLE_SECONDS)); then
      send_alert "$line"
      echo "$NOW" >"$LAST_ALERT_FILE"
    fi
  fi
done
