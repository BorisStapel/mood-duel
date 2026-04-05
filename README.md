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

2. **Set your Project ID variable**:
   ```bash
   export PROJECT_ID="mood-duel-app"
   gcloud config set project $PROJECT_ID
   ```

3. **Enable required APIs**:
   ```bash
   gcloud services enable run.googleapis.com artifact-registry.googleapis.com iam.googleapis.com cloudresourcemanager.googleapis.com
   ```

4. **Create the runtime service account** (runs the app):
   ```bash
   gcloud iam service-accounts create mood-duel-service --display-name="Mood Duel Service"
   
   # Grant minimal permissions (logging only)
   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member="serviceAccount:mood-duel-service@$PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/logging.logWriter"
   ```

5. **Create the CI/CD Deployer account**:
   ```bash
   gcloud iam service-accounts create mood-duel-deployer --display-name="Mood Duel Deployer"
   
   # Grant permissions for Cloud Run deployments
   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member="serviceAccount:mood-duel-deployer@$PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/run.admin"
   
   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member="serviceAccount:mood-duel-deployer@$PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/artifactregistry.admin"

   # Allow the deployer to "act as" the runtime service account
   gcloud iam service-accounts add-iam-policy-binding \
     mood-duel-service@$PROJECT_ID.iam.gserviceaccount.com \
     --member="serviceAccount:mood-duel-deployer@$PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/iam.serviceAccountUser"
   ```

6. **Download the CI/CD SA key** as JSON:
   ```bash
   gcloud iam service-accounts keys create ci-key.json \
     --iam-account=mood-duel-deployer@$PROJECT_ID.iam.gserviceaccount.com
   ```

7. **Add GitHub Secrets** in your repo → Settings → Secrets:
   | Secret | Value |
   |---|---|
   | `GCP_PROJECT_ID` | Your GCP project ID |
   | `GCP_SA_KEY` | The full JSON content of ci-key.json |

8. **Push to `main`** — GitHub Actions will build and deploy automatically!
   - Cloud Run instances use the least-privilege `mood-duel-service` account

### Manual deploy (without CI/CD)

```bash
# Set your project
gcloud config set project $PROJECT_ID

# Build and push image
docker build -t europe-west1-docker.pkg.dev/$PROJECT_ID/mood-duel/mood-duel:manual .
docker push europe-west1-docker.pkg.dev/$PROJECT_ID/mood-duel/mood-duel:manual

# Deploy
gcloud run deploy mood-duel \
  --image europe-west1-docker.pkg.dev/$PROJECT_ID/mood-duel/mood-duel:manual \
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
