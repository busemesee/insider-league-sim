# Insider League Simulation

Basit bir futbol ligi simülasyonu; Go backend, HTML/JS frontend ve Postman koleksiyonu içerir.

## Gereksinimler

- Docker & Docker Compose
- (Opsiyonel) Python3 (frontend’i statik sunmak için)

## Çalıştırma

Docker Compose ile ayağa kaldır:
```bash
docker compose up -d --build
## Frontend (opsiyonel)

Statik dosyayı yerelde görmek istersen:
```bash
cd frontend
python3 -m http.server 8000
