FROM python:3.11-slim

RUN apt-get update && apt-get install -y \
    curl build-essential ffmpeg && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY worker/requirements.txt /app/worker/requirements.txt
RUN pip install --no-cache-dir -r /app/worker/requirements.txt

COPY worker/ /app/worker/
COPY shared/ /app/shared/

CMD ["python", "worker/main.py"]