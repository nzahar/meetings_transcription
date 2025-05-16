import asyncio
from datetime import datetime
from dotenv import load_dotenv
import os
import sys
import logging

import database
import config
import aiohttp
from aiogram import Bot, \
    Dispatcher
from aiogram.enums.parse_mode import ParseMode
from aiogram.fsm.storage.memory import MemoryStorage
from aiogram.client.bot import DefaultBotProperties
from aiogram.utils.chat_action import ChatActionMiddleware

from handlers.handlers import router
from kb.kb import set_commands

#logging.basicConfig(level=logging.INFO, filename="py_log.log", filemode="w")
logging.basicConfig(level=logging.INFO)

load_dotenv()
telegram_token = os.getenv('TELEGRAM_TOKEN')
config.comradeai_token = os.getenv('COMRADE_TOKEN')
if not telegram_token:
    logging.error("Telegram token not set in env variables")
    sys.exit(1)
config.bot = Bot(token=telegram_token, default=DefaultBotProperties(parse_mode=ParseMode.HTML))


async def main():
    logging.info(str(datetime.now()) + " Starting bot...")
    database.db.connect()

    session = aiohttp.ClientSession()  # Создаем ClientSession

    try:
        dp = Dispatcher(storage=MemoryStorage())
        dp.include_router(router)
        dp.startup.register(set_commands)
        dp.message.middleware(ChatActionMiddleware())

        # Запускаем задачи асинхронно
        bot_task = asyncio.create_task(dp.start_polling(config.bot, allowed_updates=dp.resolve_used_update_types(), session=session))

        try:
            await asyncio.gather(bot_task)
        except Exception as ex:
            print("Error executing concurrent threads: " + str(ex))
            sys.exit(1) # Or handle the exception more gracefully
    except Exception as ex:
        logging.error(f"Main fell with error: {ex}")
    finally:
        await session.close() # Ensure the session is closed
        logging.info("Quit...")
        await config.bot.delete_webhook(drop_pending_updates=True)  # Удаляем вебхук и все обновления


if __name__ == "__main__":
    asyncio.run(main())