FROM node:20
WORKDIR /app
COPY . .
RUN npm install && npm install -D tailwindcss postcss autoprefixer && npx tailwindcss init -p
CMD ["npm", "run", "dev", "--", "--host"]
