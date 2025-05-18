FROM python:3.11

RUN apt-get update && apt-get install -y \
    curl build-essential && \
    apt-get clean

WORKDIR /app
COPY bot/ /app/bot/
COPY shared/ /app/shared/

RUN pip install --upgrade pip && \
    pip install --no-cache-dir -r bot/requirements.txt

COPY .env /app/.env

CMD [ "python", "bot/main.py" ]
