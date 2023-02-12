# https://go-social-media-js.herokuapp.com

This is my Golang & React social media personal project. I made tests for the React components. A few of them are broken, but most pass. This is my 2nd project using Golang, my 1st project was a chat app broken by non thread safe maps, dataraces and bad routing.

I tried to make the website accessible after completing it. I also wrote unit tests for a lot of the react components, 3 of the tests fail, the other 50 pass. It's definitely not accessible.

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
