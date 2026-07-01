# Barebone dev image: runs the Vite dev server with hot reload.
FROM node:22-alpine
WORKDIR /app

# Install dependencies first for better layer caching
COPY package.json package-lock.json* ./
RUN npm install

COPY . .

EXPOSE 5173
CMD ["npm", "run", "dev"]
