FROM debian:bookworm-slim

WORKDIR /app

# Instala o wget no Debian
RUN apt-get update && apt-get install -y wget && rm -rf /var/lib/apt/lists/*

# Baixa e configura o binário
RUN wget -O whats-entry-point https://github.com/viniggjpormade/whats-entry-point/releases/latest/download/whats-entry-point && \
    chmod +x whats-entry-point && \
    mv whats-entry-point /usr/local/bin/whats-entry-point

COPY .env .

ENTRYPOINT ["whats-entry-point"]