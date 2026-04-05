#!/bin/bash
# setup_gcp.sh - Automates the initial GCP infrastructure setup

set -e # Exit on error

if [ -z "$PROJECT_ID" ]; then
    echo "Error: PROJECT_ID environment variable is not set."
    exit 1
fi

echo "🚀 Starting setup for project: $PROJECT_ID"

gcloud config set project $PROJECT_ID

# Enable APIs
echo "📦 Enabling APIs..."
gcloud services enable run.googleapis.com artifact-registry.googleapis.com iam.googleapis.com cloudresourcemanager.googleapis.com

# Create Artifact Registry Repository
echo "🐳 Creating Artifact Registry repository..."
gcloud artifacts repositories create mood-duel \
    --repository-format=docker \
    --location=europe-west1 \
    --description="Docker repository for Mood Duel" || true

# Create Runtime SA
echo "👤 Creating runtime service account..."
gcloud iam service-accounts create mood-duel-service --display-name="Mood Duel Service" || true
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:mood-duel-service@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/logging.logWriter"

# Create Deployer SA
echo "🤖 Creating deployer service account..."
gcloud iam service-accounts create mood-duel-deployer --display-name="Mood Duel Deployer" || true
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:mood-duel-deployer@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/run.developer"
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:mood-duel-deployer@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

# Grant "ActAs" permission
gcloud iam service-accounts add-iam-policy-binding \
  mood-duel-service@$PROJECT_ID.iam.gserviceaccount.com \
  --member="serviceAccount:mood-duel-deployer@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

echo "✅ Setup complete!"