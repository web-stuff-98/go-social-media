This is the second version of my Go project, it has much cleaner structure and actually works 

It uses net/http, gorilla mux and gorilla websocket, you can download and upload attachments and watch the progress bar. It uploads the files in chunks, its also really slow because it's using MongoDB for storage on the free tier.

### Stuff
- Group & private video chat using simple-peer
- Nested comments
- Types & models for socket events
- Context API & Reducers
- Formik
- File attachments
- Rate limiting
- Pagination/Filtering/Sorting
- Updates & DB cleanup via changestreams
- Recursion for deleting orphans & downloading files
- Refresh tokens