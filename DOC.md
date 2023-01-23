# Server

### Socket event handling

socket.go handles all of the inbound messages from the client, using the models stored in socketModels.go. When a message comes in from the client socket.go will look for the "event_type" key and use that to determine what it needs to do.

When a socket event is sent out to a client, the client will look for the "TYPE" key to determine what it needs to do with the data. If you look inside of the frontend utils folder you will find a file called DetermineSocketEvent.ts which is used to infer the type of incoming events.

### SocketServer & AttachmentServer
SocketServer handles storing data related to users & their socket connections, stores subscriptions, and contains channels for sending data. It contains a cleanup ticker which clears its memory when subscriptions are empty.

AttachmentServer works with the attachment API handler to keep track of uploads and send status updates using the SocketServer.

### Changestreams


### Handler dependency injection

SocketServer, AttachmentServer, and database collections are injected into all of the API route handlers, from the handler.go file.

# Frontend

The frontend is built around React, Reacts Context API using Reducers, SCSS modules, flex containers, Formik, Axios and Zod. I use my IResMsg interface (response message : {error, message, pending}) with useEffect & setState inplace of React Query or similar.

### Context

- SocketContext controls opening and closing subscriptions, and opens an error modal when the server socket handler sends back an internal error. When the socket connection is lost, all the previous subscriptions will be opened back up again automatically.
- UsersContext stores all the data for users. cacheUserData is used by other components to cache a users data after receiving data from a request (if the users data is not stored already, it will be queried for). The User component works with UsersContext to make sure that User data is present only when needed, based on visibility.
- ChatContext contains the code for video chat, which is shared by the Conversations component and the Room component. It also handles navigating through sections, opening the room editor and opening chatrooms. ChatContext is created from the Chat component, it's not stored in the context folder.
- InterfaceContext measures the dimensions of the display, adjusts the horizontal whitespace, and controls dark mode. It also shares a boolean called "isMobile" which is used by components.
- ModalContext can be used to open up a confirmation modal, which can be passed a callback, display a message with an error, a loading spinner, or just a plain text message.
- PostsContext controls post card data, blog navigation, pagination, sorting, and filtering by getting and setting the URL params. It also watches for changes on the post_feed subscription, and opens/closes subscriptions on post_cards (where post_card=thepostid). Debouncing is also handled here, everything to do with blog posts.
- AuthContext contains all the functions related to authentication. It also queries the server to refresh the Users token.
- UserdropdownContext just contains the code for the dropdown that appears when you click on a users profile picture.
