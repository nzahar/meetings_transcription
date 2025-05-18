from enum import Enum

bot = None
db_file = "../shared/c_bot.db"

text_agent_name = "OpenAI_GPT_Completions"
text_agent_params = {"model": "gpt-41-vision"}
audio_agent_name = "Whisper_v3_Large"

max_requests_per_hour = 5

class MessageType(Enum):
    M_WHISPER = "T" # recognized text

