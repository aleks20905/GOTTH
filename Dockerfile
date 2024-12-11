# Stage 1: Base image for both dev and production
FROM golang:1.22-alpine AS base

WORKDIR /app

# Install dependencies required for the build process
RUN apk add --no-cache make curl git

# Install tailwindcss (adjust for architecture if needed)
RUN curl -L https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.4/tailwindcss-linux-x64 -o /usr/local/bin/tailwindcss \
    && chmod +x /usr/local/bin/tailwindcss

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Stage 2: Development build
FROM base AS development

WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . ./

# Set the environment variable if required for the makefile
ARG APP_NAME=myapp
ENV APP_NAME=${APP_NAME}

# Run the make build
RUN make build

# Stage 3: Production build
FROM base AS production

WORKDIR /app

# Copy built artifacts from the development stage
COPY --from=development /app /app

# Set permissions for the golang user
USER golang

CMD ["./bin/myapp"]
