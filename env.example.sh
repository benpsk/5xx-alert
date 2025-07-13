#!/bin/bash

# Email Config
export ENABLE_MAIL=
export MAIL_HOST=smtp.gmail.com
export MAIL_PORT=587
export MAIL_USERNAME=
export MAIL_PASSWORD=
export MAIL_TO=
export MAIL_SUBJECT="🚨 5xx Error Detected"

# Telegram Config
export ENABLE_TELEGRAM=
export TELEGRAM_BOT_TOKEN=
export TELEGRAM_CHAT_ID=

# Discord Config
export ENABLE_DISCORD=
export DISCORD_WEBHOOK_URL=

# Line Config
export ENABLE_LINE=true
export LINE_CHANNEL_TOKEN=eyJ...
export LINE_USER_ID=Uxxxxxxxxxxxxxxx

# Log File and Alert Control
export LOG_FILE=/var/log/nginx/error.log
export THROTTLE_SECONDS=60
