FROM node:20-slim

WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm install

# Copy application
COPY . .

# Expose port
EXPOSE 6002

# Run the application
CMD ["node", "app.js"]
