import asyncio
import sys

from aiogram.types import Message as AMessage

from ComradeAI.Mycelium import Mycelium, Dialog, Agent

import logging
from datetime import datetime

from aiogram.exceptions import TelegramBadRequest

import textwrap
from typing import Union
from shared import database
from bot import config
import text
from bot.config import MessageType


def register_user(user_id: int, user_name: str) -> Union[bool, str]:
    try:
        user, created = database.Users.get_or_create(user_id=user_id, defaults={"user_name": user_name})
        return True
    except Exception as e:
        return f"register_user - Произошла ошибка при работе с базой данных: {e}"


async def send_long_message(chat_id, text, keyboard=None, max_length=4096):
    print(str(datetime.now()) + " User chat id: " + str(chat_id) + " send_long_message start", flush=True)
    parts = textwrap.wrap(text, max_length)
    for i, part in enumerate(parts):
        if i == len(parts) - 1:
            if keyboard:
                await config.bot.send_message(chat_id, part, reply_markup=keyboard)
            else:
                await config.bot.send_message(chat_id, part)
        else:
            await config.bot.send_message(chat_id, part)
    print(str(datetime.now()) + " User chat id: " + str(chat_id) + " send_long_message done", flush=True)


async def edit_message(chat_id: int, message_id: int, new_text: str):
    try:
        await config.bot.edit_message_text(
            chat_id=chat_id,
            message_id=message_id,
            text=new_text
        )
    except TelegramBadRequest as e:
        logging.error(f"Failed to edit message: {e}")


async def process_audio(message: AMessage, file_id: str, mime_type: str):
    try:
        file_info = await config.bot.get_file(file_id)
        downloaded_file = await config.bot.download_file(file_info.file_path)
        audio_bytes = downloaded_file.read()

        reply_message = await message.reply(text.gen_wait)
        reply_message_id = reply_message.message_id

        asyncio.create_task(read_audio(message.from_user.id, reply_message_id, audio_bytes,
                                             mime_type))
        return None
    except Exception as ex:
        sys.stdout.flush()
        await message.answer(f"Your message caused an error: {ex}")
        return False


async def read_audio(user_id: int, message_id: int, file_data, f_type):
    logging.info(f"read_audio: START")
    task_date_start = datetime.now()
    myceliumRouter = Mycelium(host=config.comradeai_host, ComradeAIToken=config.comradeai_token, dialogs={})
    #myceliumRouter.connect()
    audio_agent = Agent(myceliumRouter, config.audio_agent_name)
    result_dialog = await asyncio.to_thread(
        lambda: Dialog.Create(audioPrompt=file_data, audioMimeType=f_type) >> audio_agent)
    answer = result_dialog.messages[-1].unified_prompts[-1].content
    logging.info(f"read_audio: FINISH")
    task_date_end = datetime.now()
    total_time = task_date_end - task_date_start
    print(f"Затраченное время: {total_time.total_seconds():.4f} секунд")

    if type(answer) is str:
        await messages_handler(MessageType.M_WHISPER.value, user_id, message_id, answer)
    else:
        if message_id is not None:
            await edit_message(int(user_id), message_id, text.error_during_execute)
        else:
            await send_long_message(int(user_id), text.error_during_execute)


async def make_things_nice(user_id: int, message_id: int, s_text: str):
    logging.info(f"Читаю текст, полученный из аудио")
    s_text = str(text.make_things_nice_prompt_begin) + " " + s_text
    myceliumRouter = Mycelium(ComradeAIToken=config.comradeai_token, dialogs={})
    myceliumRouter.connect()

    text_agent = Agent(myceliumRouter, config.text_agent_name, config.text_agent_params)
    result_dialog = await asyncio.to_thread(
        lambda: s_text >> text_agent)
    answer = result_dialog.messages[-1].unified_prompts[-1].content
    await messages_handler(MessageType.M_RESULT.value, user_id, message_id, answer)


async def messages_handler(message_type: MessageType, user_id: int, message_id: int, result_text: str):
    logging.info(f"messages_handler got message: Type={message_type} Str={result_text}")
    if result_text:
        if message_type == config.MessageType.M_WHISPER.value:
            if message_id is not None:
                await edit_message(int(user_id), message_id, text.start_getting_params)
            # отправляем текст на причесывание
            await make_things_nice(user_id, message_id, result_text)
        else:
            logging.info("-> result to user")
            await send_long_message(int(user_id), result_text)
    else:
        logging.warning(f" Received message has Null result_text")

