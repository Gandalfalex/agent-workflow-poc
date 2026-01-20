FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build
RUN chmod +x scripts/run-skill.js scripts/skill-cli.sh
EXPOSE 9000
CMD ["node", "dist/index.js"]
