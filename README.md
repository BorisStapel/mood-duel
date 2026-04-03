# ⚔️ Mood Duel

A real-time two-player emoji battle game built with Go and WebSockets.  
Two players join a room, pick emoji "moves" each round, and a rule engine decides who wins — with flavourful battle commentary.

## 🎮 How to Play

1. Open the app in two browser tabs (or share the URL with a friend)
2. Enter your name and a room code (or generate a random one)
3. Both players pick an emoji each round
4. The rule engine resolves the clash — first to dominate wins bragging rights!

---

## 🛠 Local Development

### Prerequisites
- [Go 1.21+](https://go.dev/dl/)

### Run locally

```bash
# Clone the repo
git clone https://github.com/YOUR_USERNAME/mood-duel.git
cd mood-duel

# Download dependencies
go mod download

# Run the server
go run .
```

Then open [http://localhost:8080](http://localhost:8080) in two tabs.

---

## 🐳 Run with Docker

```bash
docker build -t mood-duel .
docker run -p 8080:8080 mood-duel
```

---

## ☁️ Deploy to Google Cloud Run

### One-time setup

1. **Create a GCP project** at [console.cloud.google.com](https://console.cloud.google.com)

2. **Enable APIs**:
   ```bash
   gcloud services enable run.googleapis.com containerregistry.googleapis.com iam.googleapis.com
   ```

3. **Create a CI/CD Service Account** with deployment permissions:
   ```bash
   gcloud iam service-accounts create mood-duel-ci --display-name="Mood Duel CI/CD"
   
   # Grant permissions for Cloud Run deployments
   gcloud projects add-iam-policy-binding PROJECT_ID \
     --member="serviceAccount:mood-duel-ci@PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/run.admin"
   
   gcloud projects add-iam-policy-binding PROJECT_ID \
     --member="serviceAccount:mood-duel-ci@PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/storage.admin"
   
   gcloud projects add-iam-policy-binding PROJECT_ID \
     --member="serviceAccount:mood-duel-ci@PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/iam.serviceAccountUser"
   ```

4. **Download the CI/CD SA key** as JSON:
   ```bash
   gcloud iam service-accounts keys create ci-key.json \
     --iam-account=mood-duel-ci@PROJECT_ID.iam.gserviceaccount.com
   ```

5. **Add GitHub Secrets** in your repo → Settings → Secrets:
   | Secret | Value |
   |---|---|
   | `GCP_PROJECT_ID` | Your GCP project ID |
   | `GCP_SA_KEY` | The full JSON content of ci-key.json |

6. **Push to `main`** — GitHub Actions will build and deploy automatically!
   - The CI/CD creates a dedicated `mood-duel-service` account with minimal permissions (logging only)
   - Cloud Run instances use this least-privilege service account

### Manual deploy (without CI/CD)

```bash
# Set your project
gcloud config set project YOUR_PROJECT_ID

# Build and push image
docker build -t gcr.io/YOUR_PROJECT_ID/mood-duel .
docker push gcr.io/YOUR_PROJECT_ID/mood-duel

# Deploy
gcloud run deploy mood-duel \
  --image gcr.io/YOUR_PROJECT_ID/mood-duel \
  --platform managed \
  --region europe-west1 \
  --allow-unauthenticated \
  --port 8080
```

---

## 🧠 Project Structure

```
mood-duel/
├── main.go                    # HTTP server, WebSocket hub, game logic
├── static/
│   └── index.html             # Full frontend (HTML + CSS + JS)
├── Dockerfile                 # Multi-stage Docker build
├── .github/workflows/
│   └── deploy.yml             # CI/CD to Cloud Run on push to main
├── go.mod / go.sum
└── README.md
```

## 🔧 Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go (stdlib + gorilla/websocket) |
| Frontend | Vanilla HTML / CSS / JS |
| Real-time | WebSockets |
| Container | Docker (multi-stage) |
| Cloud | Google Cloud Run |
| CI/CD | GitHub Actions |

---

Made with ☕ and emoji combat energy.
