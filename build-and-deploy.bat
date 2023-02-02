@echo off
cd frontend
npm run build
move build ..\server
cd server
docker build -t go-social-media-js .
docker run go-social-media-js