from enum import Enum
bot = None
db_file = "c_bot.db"
date_format = "%Y-%m-%d %H:%M:%S"

comradeai_token = ""
comradeai_host = "5.35.11.211"

#text_agent_name = "Alibaba_Qwen25_Instruct"
text_agent_name = "OpenAI_GPT_Completions"
text_agent_params = {"model": "gpt-41-vision"}
text_agent = None
audio_agent_name = "Whisper_v3_Large"
#audio_agent_name = "whisper-1"
audio_agent = None

max_requests_per_hour = 5

max_retries = 10
retry_delay = 5

class States(Enum):
    S_START = "start"
    S_ENTER_DATA = "entering_data"
    S_DONE = "done"

class MessageType(Enum):
    M_WHISPER = "T" # recognized text
    M_CHAT = "D" # recognized data
    M_FILE = "F" # file to get data from
    M_COMPLAINS = "C" # recognized text to search complains in
    M_RESULT = "R"
    M_ERROR = "E"
