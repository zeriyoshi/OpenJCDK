# Alpine is not the best option. But it is good enough.
FROM python:3.13.1-alpine3.21

ENV PYTHONUNBUFFERED 1

COPY ./bot.py /application/bot.py
COPY ./requirements.txt /application/requirements.txt

WORKDIR /application

RUN pip install --no-cache-dir -r "requirements.txt"

CMD /usr/local/bin/python "/application/bot.py"
