FROM python:3.7-alpine
EXPOSE 8080
WORKDIR /opt/code
RUN pip install aiohttp==3.5.4
COPY . .
CMD ["python", "app.py"]
