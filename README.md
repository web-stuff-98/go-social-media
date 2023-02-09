# https://go-social-media-js.herokuapp.com

This is my Golang & React social media personal project. I made tests for the React components. A few of them are broken, but most pass.

Pagination count does not update properly when filtering by tags or search term. I will fix that soon.

## Features

- Private & group video chat
- Filesharing
- Live embedded comments
- Live blog
- Live everything else
- Aria labelling & tab indexes
- Notifications
- Member invites & bans
- Private & public rooms
- Customizable rooms
- Lazy loaded images with placeholders
- Rich text editor
- Mobile responsive
- Dark mode
- Live edit & delete

# Frontend

I built the frontend around Reacts Context API & SCSS. I prefer installing as few libraries as possible and using bare React for state management and data fetching.

## Frontend stuff

- Lazy loading & intersection observers
- Search, sort & pagination using URL params
- Interfaces for socket events
- useCallback, useMemo, useTransition & useDeferredValue
- AbortControllers
- WebRTC video chat using simple-peer
- Jest & RTL unit tests
- Formik & Zod
- Debouncer on search functions
- Mobile responsive layout
- Custom components for re-use of code

# Server

I used the latest version of Go for the backend, and hosted it using Heroku and Docker. I used net/http, gorilla mux router and gorilla websocket. I used MongoDB as the database and Redis for storing rate limiter data. I started with Gorm and postgres but changed my mind when I realized that doing cascading deletes and relationships is extraordinarily complicated compared to using MongoDB and changestreams. I didn't write any tests for the server because I was more interested in just using Go to make something rather than writing boring tests, but I did run it with the -race option to check for data races and resolve them.

## Server stuff

- Mutex locks & channels for thread safe maps
- Private & public socket subscriptions
- MongoDB changestreams
- Rate limiting using Redis
- Chunked octet-stream file uploads & downloads
- Automatic cleanup of DB
- Automatic cleanup of maps
- Recursive functions for deleting comment chains & chunks
- HTTP only secure logins with sessions & refresh tokens
