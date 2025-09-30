# Multi-stage build for React app with Vite
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Clear npm cache and install dependencies
RUN npm cache clean --force && \
    npm install --legacy-peer-deps

# Copy source code
COPY . .

# Set build-time environment variables
ARG REACT_APP_API_URL=https://xcxigmymtw.us-east-1.awsapprunner.com
ARG REACT_APP_WS_URL=wss://xcxigmymtw.us-east-1.awsapprunner.com/ws
ENV REACT_APP_API_URL=$REACT_APP_API_URL
ENV REACT_APP_WS_URL=$REACT_APP_WS_URL

# Build the application
RUN npm run build

# Production stage
FROM nginx:alpine AS frontend-production

# Copy built assets from builder stage
COPY --from=frontend-builder /app/dist /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]