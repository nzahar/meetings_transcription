FROM python:3.11-slim

RUN apt-get update && apt-get install -y \
    curl build-essential && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY bot/requirements.txt /app/bot/requirements.txt
RUN pip install --no-cache-dir -r /app/bot/requirements.txt

COPY bot/ /app/bot/
COPY shared/ /app/shared/

CMD ["python", "bot/main.py"]