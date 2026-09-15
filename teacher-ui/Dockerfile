FROM node:18-alpine as builder

ARG BACKEND_BASE_URL=http://localhost:8080
ENV BACKEND_BASE_URL=$BACKEND_BASE_URL

# Установка рабочей директории
WORKDIR /app

# Копирование package.json и package-lock.json
COPY package*.json ./

# Установка зависимостей
RUN npm install

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN npm run build

# Финальный образ
FROM nginx:alpine

# Копирование конфигурации nginx
COPY nginx.conf /etc/nginx/nginx.conf

# Копирование собранных файлов из первого образа
COPY --from=builder /app/dist /usr/share/nginx/html

# Экспорт порта
EXPOSE 80

# Запуск nginx
CMD ["nginx", "-g", "daemon off;"]