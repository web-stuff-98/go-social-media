This is the newest version of my Go project, it has much cleaner structure. It uses net/http, gorilla mux and gorilla websocket, and it has blogging with embedded comments, it isn't broken and theres a way of easily subscribing to stuff using the SocketServer struct.

It is not done. I need to add chatrooms, improve the UI, add rate limiters and maybe replace the auth check in the API routes with middleware. Also some optimizations to do with subscriptions.

Blogging and commenting is finished but needs optimization probably.