# https://go-social-media-js.herokuapp.com

pSQL-Social is the "new" version of this, using Vue. This project uses React 18 and postgres. The server code is also messy and has lots of nested for loops

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
