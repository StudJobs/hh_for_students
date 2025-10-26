FROM python:3.12-slim

WORKDIR /code

ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1

COPY . /code

EXPOSE 8001

RUN pip install --no-cache-dir --upgrade -r requirements.txt
