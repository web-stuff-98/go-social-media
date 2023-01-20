This is the second version of my Go project, it has much cleaner structure and actually works properly.

It uses net/http, gorilla mux and gorilla websocket, you can download and upload attachments and watch the progress bar. It uploads the files in chunks, its also really slow because it's using MongoDB for storage on the free tier.

I wasted about 1 week trying to do uploads over websockets and getting video attachment streaming to work. The code for the video playback API route is still there but commented out. You can download videos instead of watching them in the browser, uploads are handled with a HTTP endpoint that takes in chunks of binary data.

### Main features
- Nested comments
- Live updates for everything
- File attachments
- Rate limiting
- Pagination/Filtering/Sorting
- Updates & DB cleanup via changestreams
- Recursion for deleting orphans & downloading files
- Refresh tokens
