# Publish the demo safely

This project currently inherits a Git repository rooted at C:\Users\agsei.
Do not run git add . there: it could collect unrelated files from your home
folder. Create or clone a dedicated repository for this project first.

## Option A: create a dedicated repository

Copy the project to a clean location outside another Git repository, for example
C:\Projects\goFM-server. Then run these commands in PowerShell:

git init
git add .
git commit -m "OpenBridge demo: routes, SQLite storage, coding help"
git branch -M main
git remote add origin https://github.com/YOUR-ACCOUNT/YOUR-REPOSITORY.git
git push -u origin main

## Option B: clone the destination first

From C:\Projects, run git clone https://github.com/YOUR-ACCOUNT/YOUR-REPOSITORY.git goFM-server.
Copy this project into the cloned folder, excluding config.json, data, logs,
bin, web/node_modules, and web/dist. Then run these commands in the cloned
folder:

git status
git add README.md config.example.json go.mod go.sum cmd internal docs web .gitignore
git commit -m "OpenBridge demo: routes, SQLite storage, coding help"
git push origin main

## Before every push

Run go test ./..., then go vet ./.... In the web folder run npm run build.

Never add config.json, data/gofm.sqlite, logs/requests.enc, or real tokens to
GitHub.
