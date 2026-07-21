FROM python:3.12-alpine
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY main.py test_main.py ./
RUN python -c "import test_main; test_main.test_demo_marks_mutating_tools()"
EXPOSE 8080
CMD ["python", "main.py"]
