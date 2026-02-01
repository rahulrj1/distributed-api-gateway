FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application
COPY . .

# Expose port
EXPOSE 6000

# Run with gunicorn for production
CMD ["gunicorn", "--bind", "0.0.0.0:6000", "--workers", "2", "app:app"]
