from aiogram import F, Router
from aiogram.filters import Command
from aiogram.types import Message as AMessage, ContentType

from datetime import datetime
import sys
import logging

import text
import config
import utils
from shared.database import is_user_rate_limited, log_request, cleanup_old_logs

router = Router()

@router.message(Command("start"))
async def start_handler(msg: AMessage):
    chat_id = str(msg.chat.id)
    logging.info(str(datetime.now()) + " User chat id: " + str(chat_id) + " message_handler start start")
    logging.info(str(datetime.now()) + " User  id: " + str(msg.from_user.id) + f" User name: {msg.from_user.full_name}")
    result = utils.register_user(msg.from_user.id, msg.from_user.full_name)
    if result is True:
        await msg.answer(text.greet.format(name=msg.from_user.full_name))
    else:
        logging.error(f"Ошибка регистрации пользователя: {result}")
        await msg.answer(f"Ошибка регистрации пользователя: {result}")


ALLOWED_AUDIO_TYPES = {"audio/mpeg", "audio/ogg", "audio/wav", "audio/mp3"}


@router.message(F.content_type.in_({ContentType.AUDIO, ContentType.VOICE}))
async def handle_media_message(message: AMessage):
    if is_user_rate_limited(message.from_user.id, config.max_requests_per_hour):
        await message.reply(
            f"Вы превысили лимит запросов ({config.max_requests_per_hour} в час), попробуйте позже."
        )
        return

    logging.info(f"{datetime.now()} User dialog_id: {message.chat.id} Receiving voice from user")
    sys.stdout.flush()

    if message.voice:
        await utils.process_audio(message, message.voice.file_id, "audio/ogg")
    elif message.audio and message.audio.mime_type in ALLOWED_AUDIO_TYPES:
        logging.info(f"message.audio.mime_type = {message.audio.mime_type}")
        await utils.process_audio(message, message.audio.file_id, message.audio.mime_type)

    log_request(message.from_user.id)
    cleanup_old_logs()

