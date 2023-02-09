# https://go-social-media-js.herokuapp.com

This is my Golang & React social media personal project. I made tests for the React components. A few of them are broken, but most pass. This is my 2nd project using Golang, my 1st project was a chat app broken by non thread safe maps, dataraces and bad routing.

I haven't finished aria-labelling, tab indexing or all the unit tests for the frontend but I put it in the list of features below anyway.

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
- Private & public customizable rooms
- Lazy loaded images with placeholders
- Rich text editor
- Mobile responsive
- Dark mode
- Live edit & delete

# Frontend

I built the frontend around Reacts Context API & SCSS. I prefer using as few libraries as possible and using bare React for state management and data fetching. I installed lodash for debouncing though, I am not sure if it's working as intended but I might uninstall lodash and replace debounce with my own debouncer if I find out it's not actually debouncing.

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
- Custom components & stylesheets for re-use (forms labels & inputs, dropdowns & toggles)

# Server

I used Go for the backend, and hosted it using Heroku and Docker. I used net/http, gorilla mux router and gorilla websocket. I used MongoDB as the database and Redis for storing rate limiter data. I started with Gorm and postgres but changed my mind when I realized that doing cascading deletes and relationships is convoluted compared to using Prisma as a SQL ORM, or just using MongoDB and changestreams, I was going to go with Prisma to begin with but the Prisma ORM for Golang is defunct. I didn't write any tests for the Go server because I was more interested in just using Go, but I did run it with the -race option to check for data races and resolve them. I managed to get it to run on Heroku with Docker eventually by trying random commands that looked right and combining old medium articles and blog posts.

## Server stuff

- Mutex locks & channels for thread safe maps in AttachmentServer & SocketServer
- Private & public socket subscriptions
- Chunked octet-stream file uploads & downloads
- Automatic cleanup of DB
- Automatic cleanup of maps
- MongoDB changestreams
- Rate limiting using Redis
- Recursive functions for deleting comment chains & chunks
- HTTP only secure logins with sessions & refresh tokens
