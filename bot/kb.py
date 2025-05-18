from aiogram.types import BotCommand, BotCommandScopeDefault
from text import command_start


async def set_commands(bot):
    commands = [BotCommand(command='start', description=command_start)]
    await bot.set_my_commands(commands, BotCommandScopeDefault())
