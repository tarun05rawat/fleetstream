# Multi-stage build for React app with Vite
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# Copy package files
COPY package*.json ./

RUN npm cache clean --force && npm install --legacy-peer-deps

# Copy all source (including .env.production if present)
COPY . .

# ✅ Use VITE_ instead of REACT_APP_
ARG VITE_API_URL=http://factoryflow-backend-env.eba-pq48ug4g.us-east-1.elasticbeanstalk.com
ARG VITE_WS_URL=ws://factoryflow-backend-env.eba-pq48ug4g.us-east-1.elasticbeanstalk.com/ws
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_WS_URL=$VITE_WS_URL

# Build the application with env injected
RUN npm run build

# Production stage
FROM nginx:alpine AS frontend-production

# Copy built assets from builder stage
COPY --from=frontend-builder /app/dist /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
