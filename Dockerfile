# Alpine is not the best option. But it is good enough.
FROM python:3.12-alpine

ENV PYTHONUNBUFFERED True

COPY . ./

RUN pip install --no-cache-dir -r "requirements.txt"

CMD /usr/local/bin/python "bot.py"
